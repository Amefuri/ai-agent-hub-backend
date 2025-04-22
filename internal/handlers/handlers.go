package handlers

import (
	"ai-agent-hub/internal/models"
	"ai-agent-hub/internal/utils"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// ========== AUTH ==========

// POST /api/auth/register
func (h *Handler) Register(c echo.Context) error {
	type RegisterRequest struct {
		Username string `json:"username" validate:"required,min=3,max=32"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	var req RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Check for existing user
	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err != gorm.ErrRecordNotFound {
		return c.JSON(http.StatusConflict, echo.Map{"error": "User already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to hash password"})
	}

	// Save user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "User registered successfully"})
}

// POST /api/auth/login
func (h *Handler) Login(c echo.Context) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Database error"})
	}

	// Compare password
	if err := user.CheckPassword(req.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	// Create JWT
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // 3 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not sign token"})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": signedToken})
}

// ========== AGENT CRUD ==========

// GET /api/agents
func (h *Handler) GetAgents(c echo.Context) error {

	p := utils.GetPagination(c)

	var agents []models.Agent
	var total int64

	h.DB.Model(&models.Agent{}).Count(&total)
	if err := h.DB.Limit(p.Limit).Offset(p.Offset).Find(&agents).Error; err != nil {
		return err
	}

	resp := utils.NewPaginatedResponse(agents, p.Page, p.Limit, total)
	return c.JSON(http.StatusOK, resp)
}

// GET /api/agents/:id
func (h *Handler) GetAgentsByID(c echo.Context) error {
	id := c.Param("id")

	var agent models.Agent
	if err := h.DB.First(&agent, "id = ?", id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Agent not found")
	}

	return c.JSON(http.StatusOK, agent)
}

// GET /api/user/:user_id/agents
func (h *Handler) GetAgentsOfUserID(c echo.Context) error {
	p := utils.GetPagination(c)
	userID := c.Param("user_id")

	var agents []models.Agent
	var total int64

	h.DB.Model(&models.Agent{}).Count(&total)
	if err := h.DB.Where("user_id = ?", userID).Find(&agents).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch agents for user")
	}

	resp := utils.NewPaginatedResponse(agents, p.Page, p.Limit, total)
	return c.JSON(http.StatusOK, resp)
}

//===============================================================================================================

// GET /api/my/agents
func (h *Handler) GetMyAgents(c echo.Context) error {
	p := utils.GetPagination(c)

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var agents []models.Agent
	var total int64

	h.DB.Model(&models.Agent{}).Count(&total)
	if err := h.DB.Where("user_id = ?", userID).Find(&agents).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch your agents")
	}

	resp := utils.NewPaginatedResponse(agents, p.Page, p.Limit, total)
	return c.JSON(http.StatusOK, resp)
}

// GET /api/my/agents/:id
func (h *Handler) GetMyAgentByID(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	agentID := c.Param("id")

	var agent models.Agent
	if err := h.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Agent not found"})
		}
		return err
	}

	return c.JSON(http.StatusOK, agent)
}

// POST /api/my/agents
func (h *Handler) CreateMyAgents(c echo.Context) error {
	user := c.Get("user")
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing JWT token")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	// user := c.Get("user").(*jwt.Token)
	// claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64)) // JWT stores numbers as float64

	var input models.Agent
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	agent := models.Agent{
		Name:          input.Name,
		Description:   input.Description,
		Avatar:        input.Avatar,
		SystemPrompt:  input.SystemPrompt,
		InputTemplate: input.InputTemplate,
		Personality:   input.Personality,
		UserID:        userID,
	}

	if err := h.DB.Create(&agent).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create agent"})
	}

	return c.JSON(http.StatusCreated, agent)
}

// PUT /api/my/agents/:id
func (h *Handler) UpdateMyAgent(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	agentID := c.Param("id")

	var agent models.Agent
	if err := h.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Agent not found"})
	}

	var input models.Agent
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	agent.Name = input.Name
	agent.Description = input.Description
	agent.Avatar = input.Avatar
	agent.SystemPrompt = input.SystemPrompt
	agent.InputTemplate = input.InputTemplate
	agent.Personality = input.Personality

	if err := h.DB.Save(&agent).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update agent"})
	}

	return c.JSON(http.StatusOK, agent)
}

// DELETE /api/my/agents/:id
func (h *Handler) DeleteMyAgent(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	agentID := c.Param("id")

	if err := h.DB.Where("id = ? AND user_id = ?", agentID, userID).Delete(&models.Agent{}).Error; err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// // ========== CHAT ==========

// func (h *Handler) ChatWithAgent(c echo.Context) error {
// 	id := c.Param("id")
// 	var agent models.Agent
// 	if err := h.DB.First(&agent, id).Error; err != nil {
// 		return echo.ErrNotFound
// 	}

// 	var req struct {
// 		Message string `json:"message"`
// 	}
// 	if err := c.Bind(&req); err != nil {
// 		return err
// 	}

// 	// Simulate response (you'll connect to Python microservice here)
// 	response := "This is a simulated reply from agent: " + agent.Name
// 	return c.JSON(http.StatusOK, echo.Map{
// 		"agent_id": agent.ID,
// 		"message":  req.Message,
// 		"reply":    response,
// 	})
// }

// // ========== PROFILE ==========

// func (h *Handler) GetProfile(c echo.Context) error {
// 	userID := c.Get("user").(int)
// 	var user models.User
// 	if err := h.DB.First(&user, userID).Error; err != nil {
// 		return echo.ErrUnauthorized
// 	}
// 	user.Password = "" // don't return hashed password
// 	return c.JSON(http.StatusOK, user)
// }

// func (h *Handler) UpdateProfile(c echo.Context) error {
// 	userID := c.Get("user").(int)
// 	var user models.User
// 	if err := h.DB.First(&user, userID).Error; err != nil {
// 		return echo.ErrUnauthorized
// 	}

// 	var req models.User
// 	if err := c.Bind(&req); err != nil {
// 		return err
// 	}
// 	user.Username = req.Username
// 	if req.Password != "" {
// 		user.Password, _ = utils.HashPassword(req.Password)
// 	}

// 	if err := h.DB.Save(&user).Error; err != nil {
// 		return err
// 	}
// 	user.Password = ""
// 	return c.JSON(http.StatusOK, user)
// }

// // ========== LINE LINK ==========

// func (h *Handler) LinkAgentToLine(c echo.Context) error {
// 	id := c.Param("id")
// 	userID := c.Get("user").(int)

// 	var agent models.Agent
// 	if err := h.DB.First(&agent, id).Error; err != nil {
// 		return echo.ErrNotFound
// 	}
// 	if agent.UserID != uint(userID) {
// 		return echo.ErrUnauthorized
// 	}

// 	var req struct {
// 		WebhookURL string `json:"webhook_url"`
// 	}
// 	if err := c.Bind(&req); err != nil {
// 		return err
// 	}
// 	agent.LineWebhook = req.WebhookURL
// 	if err := h.DB.Save(&agent).Error; err != nil {
// 		return err
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{"message": "Agent linked to LINE", "webhook": agent.LineWebhook})
// }

//==========================================================

// package handlers

// import (
//     "net/http"
//     "github.com/labstack/echo/v4"
//     "gorm.io/gorm"
// )

// func Login(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Login"})
//     }
// }

// func Register(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Register"})
//     }
// }

// func GetAgents(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, []string{"Agent A", "Agent B"})
//     }
// }

// func GetAgentByID(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
//     }
// }

// func CreateAgent(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Create Agent"})
//     }
// }

// func UpdateAgent(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Update Agent"})
//     }
// }

// func DeleteAgent(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Delete Agent"})
//     }
// }

// func ChatWithAgent(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Chat with Agent"})
//     }
// }

// func LinkAgentLine(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Link Agent to LINE"})
//     }
// }

// func GetProfile(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Get Profile"})
//     }
// }

// func UpdateProfile(db *gorm.DB) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, map[string]string{"message": "Update Profile"})
//     }
// }
