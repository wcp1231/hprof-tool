package web

import (
	"github.com/labstack/echo/v4"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/snapshot"
	"strconv"
)

type WebEndpoint struct {
	s *snapshot.Snapshot
	e *echo.Echo
}

func NewWebEndpoint(s *snapshot.Snapshot) *WebEndpoint {
	e := echo.New()
	return &WebEndpoint{
		s, e,
	}
}

func (w *WebEndpoint) initEndpoints() {
	w.e.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})
	g := w.e.Group("/api")
	g.GET("/threads", func(c echo.Context) error {
		threads := w.s.GetThreads()
		return c.JSON(200, threads)
	})
	g.GET("/classes", func(c echo.Context) error {
		classes, err := w.s.ListClassesStatistics()
		if err != nil {
			return c.JSON(500, struct {
				Error string `json:"error"`
			}{Error: err.Error()})
		}
		return c.JSON(200, classes)
	})
	g.GET("/classes/:id/instances", func(c echo.Context) error {
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)
		typStr := c.QueryParam("type")
		typ := 0
		switch typStr {
		case "oa":
			typ = 0x22
		case "pa":
			typ = 0x23
		default:
			typ = 0x20
		}

		classes, err := w.s.ListInstancesStatistics(id, typ)
		if err != nil {
			return c.JSON(500, struct {
				Error string `json:"error"`
			}{Error: err.Error()})
		}
		return c.JSON(200, classes)
	})
	g.GET("/instances/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)

		instance, err := w.s.GetInstanceDetail(id)
		if err != nil {
			return c.JSON(500, struct {
				Error string `json:"error"`
			}{Error: err.Error()})
		}
		return c.JSON(200, instance)
	})
	g.GET("/references/:id/inbound", func(c echo.Context) error {
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)

		var result []hprof.HProfRecord
		err := w.s.GetRecordInbound(id, func(record hprof.HProfRecord) error {
			result = append(result, record)
			return nil
		})
		if err != nil {
			return c.JSON(500, struct {
				Error string `json:"error"`
			}{Error: err.Error()})
		}
		return c.JSON(200, result)
	})
}

func (w *WebEndpoint) Start(address string) {
	w.initEndpoints()
	w.e.Logger.Fatal(w.e.Start(address))
}
