package services

import (
"log"
"sync"
)

var (
once sync.Once
config ConfigService
)

// InitServices initializes all services and their dependencies
func InitServices() error {
var err error
once.Do(func() {
err = initializeServices()
})
return err
}

// GetConfig returns the singleton ConfigService instance
func GetConfig() ConfigService {
return config
}

// initializeServices sets up all services and their dependencies
func initializeServices() error {
// Initialize config service first
config = NewConfigService()

// Initialize database
if err := config.InitDB(); err != nil {
return err
}

// Initialize and register core services
services := map[string]interface{}{
"user":    NewUserService(),
"group":   NewGroupService(),
"message": NewMessageService(),
"post":    NewPostService(),
}

// Register all services with the config service
for name, service := range services {
if err := config.RegisterService(name, service); err != nil {
log.Printf("Failed to register service %s: %v", name, err)
return err
}
}

// Load environment variables and settings
if err := loadEnvironment(); err != nil {
return err
}

return nil
}

// loadEnvironment loads environment variables and application settings
func loadEnvironment() error {
// TODO: Load environment variables from .env file or environment
// For example:
// config.SetEnv("DB_HOST", os.Getenv("DB_HOST"))
// config.SetEnv("DB_PORT", os.Getenv("DB_PORT"))
// etc.

// TODO: Load application settings from configuration file or database
// For example:
// config.SetAppSetting("MAX_UPLOAD_SIZE", 10*1024*1024)
// config.SetAppSetting("ALLOWED_FILE_TYPES", []string{".jpg", ".png", ".gif"})
// etc.

return nil
}

// CleanupServices performs cleanup when shutting down the application
func CleanupServices() error {
if config != nil {
return config.CloseDB()
}
return nil
}

// Helper functions to get specific services

// GetUserService returns the UserService instance
func GetUserService() UserService {
if config == nil {
return nil
}
return config.GetUserService()
}

// GetGroupService returns the GroupService instance
func GetGroupService() GroupService {
if config == nil {
return nil
}
return config.GetGroupService()
}

// GetMessageService returns the MessageService instance
func GetMessageService() MessageService {
if config == nil {
return nil
}
return config.GetMessageService()
}

// GetPostService returns the PostService instance
func GetPostService() PostService {
if config == nil {
return nil
}
return config.GetPostService()
}
