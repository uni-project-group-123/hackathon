package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Use pure Go SQLite for cross-platform compatibility (no CGO required)
var db *gorm.DB

type Admin struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"unique" json:"username"`
	PasswordHash string `json:"-"`
}

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       string    `gorm:"unique" json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CardID       string    `json:"card_id"`
	FaceEncoding string    `json:"face_encoding"`
	CreatedAt    time.Time `json:"created_at"`
}

type Zone struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	MaxCapacity      int    `json:"max_capacity"`
	CurrentOccupancy int    `json:"current_occupancy"`
	IsRestricted     bool   `json:"is_restricted"`
}

type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`
	ZoneID    uint      `json:"zone_id"`
	Action    string    `json:"action"`
	Method    string    `json:"method"`
	Timestamp time.Time `json:"timestamp"`
	Notes     string    `json:"notes"`
	User      User      `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
	Zone      Zone      `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("sentinel.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&Admin{}, &User{}, &Zone{}, &AccessLog{})

	initDB()

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.LoadHTMLGlob("../frontend/templates/*")

	store := cookie.NewStore([]byte("sentinel-secret-key-change-in-production"))
	r.Use(sessions.Sessions("sentinel_session", store))

	r.GET("/login", loginPage)
	r.POST("/login", loginHandler)
	r.GET("/logout", logoutHandler)

	r.GET("/", authRequired(), indexPage)
	r.GET("/api/stats", authRequired(), getStats)
	r.GET("/api/access-logs", authRequired(), getAccessLogs)
	r.GET("/api/alerts", authRequired(), getAlerts)
	r.POST("/api/alerts/:id", authRequired(), updateAlert)
	r.GET("/api/users", authRequired(), getUsers)
	r.POST("/api/users", authRequired(), createUser)
	r.PUT("/api/users/:id", authRequired(), updateUser)
	r.DELETE("/api/users/:id", authRequired(), deleteUser)
	r.GET("/api/zones", authRequired(), getZones)
	r.POST("/api/zones", authRequired(), createZone)

	// Access controller endpoints (require API key)
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "face-scan-secure-key-2024"
	}
	accessControllerAuth := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			key := c.GetHeader("X-API-Key")
			if key == "" {
				key = c.PostForm("api_key")
			}
			if key != apiKey {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}
			c.Next()
		}
	}

	r.POST("/api/authenticate", accessControllerAuth(), authenticate)
	r.POST("/api/access-log", accessControllerAuth(), logAccess)
	r.GET("/api/occupancy", getOccupancy)
	r.POST("/api/door/:zone_id/:action", accessControllerAuth(), controlDoor)
	r.GET("/api/health", healthCheck)

	r.Run(":5000")
}

func initDB() {
	var admin Admin
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		admin := Admin{
			Username:     "admin",
			PasswordHash: hashPassword("admin123"),
		}
		db.Create(&admin)
	}

	var count int64
	db.Model(&Zone{}).Count(&count)
	if count == 0 {
		zones := []Zone{
			{Name: "Computer Lab 1", Description: "Main computer lab", MaxCapacity: 30},
			{Name: "Chemistry Lab", Description: "Chemistry laboratory", MaxCapacity: 20, IsRestricted: true},
			{Name: "Physics Lab", Description: "Physics laboratory", MaxCapacity: 15, IsRestricted: true},
			{Name: "Server Room", Description: "IT infrastructure room", MaxCapacity: 5, IsRestricted: true},
			{Name: "Library", Description: "Main library", MaxCapacity: 100},
		}
		db.Create(&zones)
	}

	db.Model(&User{}).Count(&count)
	if count == 0 {
		users := []User{
			{UserID: "STU001", Name: "John Doe", Email: "john@campus.edu", Role: "student", Status: "active", CardID: "STU001"},
			{UserID: "STU002", Name: "Jane Smith", Email: "jane@campus.edu", Role: "student", Status: "active", CardID: "STU002"},
			{UserID: "FAC001", Name: "Dr. Emily Brown", Email: "emily@campus.edu", Role: "faculty", Status: "active", CardID: "FAC001"},
		}
		db.Create(&users)

		logs := []AccessLog{
			{UserID: "STU001", ZoneID: 1, Action: "entry", Method: "Face Recognition", Timestamp: time.Now().Add(-2 * time.Hour)},
			{UserID: "STU002", ZoneID: 2, Action: "entry", Method: "ID Card + PIN", Timestamp: time.Now().Add(-1 * time.Hour)},
			{UserID: "STU001", ZoneID: 1, Action: "exit", Method: "Face Recognition", Timestamp: time.Now().Add(-30 * time.Minute)},
			{UserID: "FAC001", ZoneID: 3, Action: "entry", Method: "Face Recognition", Timestamp: time.Now()},
		}
		db.Create(&logs)
	}
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		loggedIn := session.Get("logged_in")
		if loggedIn == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"Error": ""})
}

func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var admin Admin
	if err := db.Where("username = ?", username).First(&admin).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid credentials"})
		return
	}

	if hashPassword(password) != admin.PasswordHash {
		c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid credentials"})
		return
	}

	session := sessions.Default(c)
	session.Set("logged_in", true)
	session.Set("username", username)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func indexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func getStats(c *gin.Context) {
	var accessLogs []AccessLog
	today := time.Now().Format("2006-01-02")
	db.Where("action != ? AND DATE(timestamp) = ?", "denied", today).Find(&accessLogs)

	var zones []Zone
	db.Find(&zones)

	var users []User
	db.Where("status = ?", "active").Find(&users)

	occupiedZones := 0
	totalOccupancy := 0
	for _, z := range zones {
		if z.CurrentOccupancy > 0 {
			occupiedZones++
		}
		totalOccupancy += z.CurrentOccupancy
	}

	c.JSON(http.StatusOK, gin.H{
		"total_access_today":     len(accessLogs),
		"active_alerts":          2,
		"zones_occupied":         occupiedZones,
		"total_users_on_campus":  totalOccupancy,
		"total_users_registered": len(users),
	})
}

func getAccessLogs(c *gin.Context) {
	var logs []AccessLog
	db.Preload("User").Preload("Zone").Order("timestamp desc").Limit(50).Find(&logs)

	result := make([]gin.H, len(logs))
	for i, log := range logs {
		result[i] = gin.H{
			"id":        log.ID,
			"user":      log.User.Name,
			"user_id":   log.UserID,
			"zone":      log.Zone.Name,
			"action":    log.Action,
			"method":    log.Method,
			"timestamp": log.Timestamp.Format("2006-01-02 15:04:05"),
		}
	}
	c.JSON(http.StatusOK, result)
}

func getAlerts(c *gin.Context) {
	alerts := []gin.H{
		{
			"id":        "1",
			"type":      "warning",
			"title":     "Unauthorized Access Attempt",
			"message":   "Student STU003 attempted to access Server Room without clearance",
			"zone":      "Server Room",
			"timestamp": "2026-03-03 09:15:33",
			"status":    "active",
		},
		{
			"id":        "2",
			"type":      "info",
			"title":     "High Occupancy",
			"message":   "Computer Lab 1 has reached 95% capacity",
			"zone":      "Computer Lab 1",
			"timestamp": "2026-03-03 09:30:00",
			"status":    "active",
		},
	}
	c.JSON(http.StatusOK, alerts)
}

func updateAlert(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getUsers(c *gin.Context) {
	var users []User
	db.Find(&users)
	c.JSON(http.StatusOK, users)
}

func createUser(c *gin.Context) {
	var input struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Role   string `json:"role"`
		CardID string `json:"card_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := User{
		UserID: input.UserID,
		Name:   input.Name,
		Email:  input.Email,
		Role:   input.Role,
		Status: "active",
		CardID: input.CardID,
	}

	db.Create(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status string `json:"status"`
		CardID string `json:"card_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	if input.Status != "" {
		user.Status = input.Status
	}
	if input.CardID != "" {
		user.CardID = input.CardID
	}

	db.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	db.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getZones(c *gin.Context) {
	var zones []Zone
	db.Find(&zones)
	c.JSON(http.StatusOK, zones)
}

func createZone(c *gin.Context) {
	var input struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		MaxCapacity  int    `json:"max_capacity"`
		IsRestricted bool   `json:"is_restricted"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone := Zone{
		Name:         input.Name,
		Description:  input.Description,
		MaxCapacity:  input.MaxCapacity,
		IsRestricted: input.IsRestricted,
	}

	db.Create(&zone)
	c.JSON(http.StatusOK, gin.H{"success": true, "zone": zone})
}

// Access Controller Endpoints

func authenticate(c *gin.Context) {
	var input struct {
		CardID    string `json:"card_id"`
		ZoneID    int    `json:"zone_id"`
		Method    string `json:"method"`
		Timestamp string `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by card ID
	var user User
	if err := db.Where("card_id = ?", input.CardID).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Card not registered",
		})
		return
	}

	// Check if user is active
	if user.Status != "active" {
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"user_id":   user.UserID,
			"user_name": user.Name,
			"message":   "User account is " + user.Status,
		})
		return
	}

	// Check zone restrictions
	var zone Zone
	if err := db.First(&zone, input.ZoneID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid zone",
		})
		return
	}

	// Faculty can access restricted zones, students need permission
	if zone.IsRestricted && user.Role == "student" {
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"user_id":   user.UserID,
			"user_name": user.Name,
			"message":   "Access denied: Restricted area",
		})
		return
	}

	// Check capacity
	if zone.CurrentOccupancy >= zone.MaxCapacity {
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"user_id":   user.UserID,
			"user_name": user.Name,
			"message":   "Zone at maximum capacity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"user_id":   user.UserID,
		"user_name": user.Name,
		"message":   "Access granted",
		"role":      user.Role,
	})
}

func logAccess(c *gin.Context) {
	var input struct {
		CardID    string `json:"card_id"`
		ZoneID    int    `json:"zone_id"`
		Method    string `json:"method"`
		Timestamp string `json:"timestamp"`
		Action    string `json:"action"`
		Success   bool   `json:"success"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by card ID
	var user User
	db.Where("card_id = ?", input.CardID).First(&user)

	// Determine action (entry/exit/denied)
	action := input.Action
	if action == "" {
		action = "denied"
	}

	// Log the access
	accessLog := AccessLog{
		UserID:    user.UserID,
		ZoneID:    uint(input.ZoneID),
		Action:    action,
		Method:    input.Method,
		Timestamp: time.Now(),
	}

	if input.Success {
		// Update zone occupancy
		var zone Zone
		if db.First(&zone, input.ZoneID).Error == nil {
			if action == "entry" {
				zone.CurrentOccupancy++
			} else if action == "exit" && zone.CurrentOccupancy > 0 {
				zone.CurrentOccupancy--
			}
			db.Save(&zone)
		}
	}

	db.Create(&accessLog)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getOccupancy(c *gin.Context) {
	var zones []Zone
	db.Find(&zones)

	occupancy := make([]gin.H, len(zones))
	for i, z := range zones {
		occupancy[i] = gin.H{
			"zone_id":           z.ID,
			"zone_name":         z.Name,
			"current_occupancy": z.CurrentOccupancy,
			"max_capacity":      z.MaxCapacity,
			"percentage":        float64(z.CurrentOccupancy) / float64(z.MaxCapacity) * 100,
		}
	}

	c.JSON(http.StatusOK, occupancy)
}

func controlDoor(c *gin.Context) {
	zoneID := c.Param("zone_id")
	action := c.Param("action")

	var zone Zone
	if err := db.First(&zone, zoneID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
		return
	}

	if action == "unlock" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Door unlocked for " + zone.Name,
		})
	} else if action == "lock" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Door locked for " + zone.Name,
		})
	} else if action == "status" {
		c.JSON(http.StatusOK, gin.H{
			"zone_id":   zone.ID,
			"zone_name": zone.Name,
			"locked":    zone.IsRestricted,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}

func healthCheck(c *gin.Context) {
	var userCount int64
	var zoneCount int64
	db.Model(&User{}).Count(&userCount)
	db.Model(&Zone{}).Count(&zoneCount)

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"users":     userCount,
		"zones":     zoneCount,
		"database":  "connected",
	})
}
