package routes

import (
	"ai-agent-hub/internal/handlers"
	"ai-agent-hub/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(e *echo.Echo, db *gorm.DB) {
	e.POST("/api/auth/login", handlers.NewHandler(db).Login)                      //Login
	e.POST("/api/auth/register", handlers.NewHandler(db).Register)                //Register
	e.GET("/api/agents", handlers.NewHandler(db).GetAgents)                       //Get Agents List
	e.GET("/api/agents/:id", handlers.NewHandler(db).GetAgentsByID)               //Get Agent Detail
	e.GET("/api/user/:user_id/agents", handlers.NewHandler(db).GetAgentsOfUserID) //Get Agents List of UserID
	e.GET("/api/agents/featured", handlers.NewHandler(db).GetFeaturedAgents)      //Get Featured Agents
	e.GET("/api/agents/popular", handlers.NewHandler(db).GetPopularAgents)        //Get Popular Agents
}

func RegisterPrivateRoutes(e *echo.Echo, db *gorm.DB) {

	// * GET /api/user/:user_id/agents: List all agents for a specific user.
	// * GET /api/user/:user_id/agents/:agent_id: View a specific agent's details for a specific user.
	// * POST /api/user/:user_id/agents: Create a new agent for a specific user (requires authentication).
	// * PUT /api/user/:user_id/agents/:agent_id: Update a user's agent (requires authentication).
	// * DELETE /api/user/:user_id/agents/:agent_id: Delete a user's agent (requires authentication).

	r := e.Group("/api/my", middleware.JWTMiddleware())
	r.GET("/agents", handlers.NewHandler(db).GetMyAgents)
	r.GET("/agents/:id", handlers.NewHandler(db).GetMyAgentByID)
	r.POST("/agents", handlers.NewHandler(db).CreateMyAgents)
	r.PUT("/agents/:id", handlers.NewHandler(db).UpdateMyAgent)
	r.DELETE("/agents/:id", handlers.NewHandler(db).DeleteMyAgent)

	// r.POST("/agents", handlers.CreateAgent(db))
	// r.PUT("/agents/:id", handlers.UpdateAgent(db))
	// r.DELETE("/agents/:id", handlers.DeleteAgent(db))
	// r.POST("/agents/:id/chat", handlers.ChatWithAgent(db))
	// r.POST("/agents/:id/line/link", handlers.LinkAgentLine(db))
	// r.GET("/profile", handlers.GetProfile(db))
	// r.PUT("/profile", handlers.UpdateProfile(db))
}
