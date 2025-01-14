package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/http/api"
	"github.com/burlingtonbertie99/mykeys/dstore"
	"github.com/burlingtonbertie99/mykeys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) sigchain(c echo.Context, kid keys.ID) (*keys.Sigchain, map[string]api.Metadata, error) {
	ctx := c.Request().Context()
	iter, err := s.fi.DocumentIterator(ctx, "sigchain", dstore.Prefix(kid.String()))
	if err != nil {
		return nil, nil, err
	}
	defer iter.Release()

	sc := keys.NewSigchain(kid)
	md := make(map[string]api.Metadata, 100)
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, nil, err
		}
		if doc == nil {
			break
		}

		var st *keys.Statement
		if err := json.Unmarshal(doc.Data(), &st); err != nil {
			return nil, nil, err
		}

		if err := sc.Add(st); err != nil {
			return nil, nil, err
		}
		md[st.URL()] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}

	return sc, md, nil
}

func (s *Server) getSigchain(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return s.ErrNotFound(c, nil)
	}

	s.logger.Infof("Loading sigchain: %s", kid)
	sc, md, err := s.sigchain(c, kid)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if sc.Length() == 0 {
		return s.ErrNotFound(c, errors.Errorf("sigchain not found"))
	}
	resp := api.SigchainResponse{
		KID:        kid,
		Statements: sc.Statements(),
	}
	fields := dstore.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getSigchainStatement(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return s.ErrNotFound(c, err)
	}
	i, err := strconv.Atoi(c.Param("seq"))
	if err != nil {
		return s.ErrNotFound(c, err)
	}
	path := dstore.Path("sigchain", kid.WithSeq(i))
	st, doc, err := s.statement(ctx, path)
	if st == nil {
		return s.ErrNotFound(c, errors.Errorf("statement not found"))
	}
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if !doc.CreatedAt.IsZero() {
		c.Response().Header().Set("CreatedAt", doc.CreatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("CreatedAt-RFC3339M", doc.CreatedAt.Format(tsutil.RFC3339Milli))
	}
	if !doc.UpdatedAt.IsZero() {
		c.Response().Header().Set("Last-Modified", doc.UpdatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("Last-Modified-RFC3339M", doc.UpdatedAt.Format(tsutil.RFC3339Milli))
	}

	return JSON(c, http.StatusOK, st)
}

func (s *Server) putSigchainStatement(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if c.Request().Body == nil {
		return s.ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	st, err := s.statementFromBytes(ctx, b)
	if err != nil {
		return s.ErrBadRequest(c, err)
	}
	if len(st.Data) > 16*1024 {
		return s.ErrBadRequest(c, errors.Errorf("too much data for sigchain statement (greater than 16KiB)"))
	}

	if c.Param("kid") != st.KID.String() {
		return s.ErrBadRequest(c, errors.Errorf("invalid kid"))
	}
	if c.Param("seq") != fmt.Sprintf("%d", st.Seq) {
		return s.ErrBadRequest(c, errors.Errorf("invalid seq"))
	}
	if st.Seq <= 0 {
		return s.ErrBadRequest(c, errors.Errorf("invalid seq"))
	}

	path := dstore.Path("sigchain", keys.StatementID(st.KID, st.Seq))

	exists, err := s.fi.Exists(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if exists {
		return s.ErrConflict(c, errors.Errorf("statement already exists"))
	}

	sc, _, err := s.sigchain(c, st.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	if sc.Length() >= 1024 {
		// TODO: Increase limits
		return s.ErrEntityTooLarge(c, errors.Errorf("sigchain limit reached, contact gabriel@github to bump the limits"))
	}

	prev := sc.Last()
	if err := sc.VerifyStatement(st, prev); err != nil {
		return s.ErrBadRequest(c, err)
	}
	if err := sc.Add(st); err != nil {
		return s.ErrBadRequest(c, err)
	}

	// Check we don't have an existing user with a different key, which would cause duplicates in search.
	// They should revoke the existing user before linking a new key.
	// Since there is a delay in indexing this won't stop a malicious user from creating duplicates but
	// it will limit them. If we find spaming this is a problem, we can get more strict.
	existing, err := s.users.CheckForExisting(ctx, sc)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if existing != "" {
		if err := s.checkKID(ctx, existing, HighPriority); err != nil {
			return s.ErrResponse(c, err)
		}
		return s.ErrConflict(c, errors.Errorf("user already exists with key %s, if you removed or revoked the previous statement you may need to wait briefly for search to update", existing))
	}

	s.logger.Infof("Statement, set %s", path)
	if err := s.fi.Create(ctx, path, dstore.Data(b)); err != nil {
		return s.ErrResponse(c, err)
	}

	if err := s.sigchains.Index(st.KID); err != nil {
		return s.ErrResponse(c, err)
	}

	if err := s.checkKID(ctx, st.KID, HighPriority); err != nil {
		return s.ErrResponse(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) statement(ctx context.Context, path string) (*keys.Statement, *dstore.Document, error) {
	e, err := s.fi.Get(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	if e == nil {
		return nil, nil, nil
	}
	st, err := s.statementFromBytes(ctx, e.Data())
	if err != nil {
		return nil, nil, err
	}
	return st, e, nil
}

func (s *Server) statementFromBytes(ctx context.Context, b []byte) (*keys.Statement, error) {
	var st *keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	bout, err := st.Bytes()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(b, bout) {
		s.logger.Errorf("%s != %s", string(b), string(bout))
		return nil, errors.Errorf("invalid statement bytes")
	}
	if err := st.Verify(); err != nil {
		return st, err
	}
	return st, nil
}

func (s *Server) getSigchainAliased(c echo.Context) error {
	if c.Request().Host == "sigcha.in" {
		return s.getSigchain(c)
	}
	return s.ErrNotFound(c, nil)

}
func (s *Server) getSigchainStatementAliased(c echo.Context) error {
	if c.Request().Host == "sigcha.in" {
		return s.getSigchainStatement(c)
	}
	return s.ErrNotFound(c, nil)
}

func (s *Server) putSigchainStatementAliased(c echo.Context) error {
	// Versions earlier <= 0.0.48 just used PUT /:kid/:seq instead of /sigchain/:kid/:seq
	// for all hosts (including keys.pub), so this change breaks them.
	if c.Request().Host == "sigcha.in" {
		return s.putSigchainStatement(c)
	}
	return s.ErrNotFound(c, nil)
}
