package services

import (
"database/sql"
"sync"
)

// ConfigService defines the interface for application configuration
type ConfigService interface {
// Database
GetDB() *sql.DB
InitDB() error
CloseDB() error

// Environment
GetEnv(key string) string
SetEnv(key, value string) error

// Application Settings
GetAppSetting(key string) (interface{}, error)
SetAppSetting(key string, value interface{}) error
GetAllSettings() (map[string]interface{}, error)

// Service Management
GetUserService() UserService
GetGroupService() GroupService
GetMessageService() MessageService
GetPostService() PostService

// Dependency Injection
RegisterService(name string, service interface{}) error
GetService(name string) (interface{}, error)
}

// ConfigServiceImpl implements the ConfigService interface
type ConfigServiceImpl struct {
mu sync.RWMutex
db *sql.DB
env map[string]string
settings map[string]interface{}
services map[string]interface{}
}

// NewConfigService creates a new ConfigService instance
func NewConfigService() ConfigService {
return &ConfigServiceImpl{
env:      make(map[string]string),
settings: make(map[string]interface{}),
services: make(map[string]interface{}),
}
}

// GetDB returns the database connection
func (s *ConfigServiceImpl) GetDB() *sql.DB {
s.mu.RLock()
defer s.mu.RUnlock()
return s.db
}

// InitDB initializes the database connection
func (s *ConfigServiceImpl) InitDB() error {
s.mu.Lock()
defer s.mu.Unlock()
// TODO: Implement database initialization
return nil
}

// CloseDB closes the database connection
func (s *ConfigServiceImpl) CloseDB() error {
s.mu.Lock()
defer s.mu.Unlock()
if s.db != nil {
return s.db.Close()
}
return nil
}

// GetEnv retrieves an environment variable
func (s *ConfigServiceImpl) GetEnv(key string) string {
s.mu.RLock()
defer s.mu.RUnlock()
return s.env[key]
}

// SetEnv sets an environment variable
func (s *ConfigServiceImpl) SetEnv(key, value string) error {
s.mu.Lock()
defer s.mu.Unlock()
s.env[key] = value
return nil
}

// GetAppSetting retrieves an application setting
func (s *ConfigServiceImpl) GetAppSetting(key string) (interface{}, error) {
s.mu.RLock()
defer s.mu.RUnlock()
if value, ok := s.settings[key]; ok {
return value, nil
}
return nil, nil
}

// SetAppSetting sets an application setting
func (s *ConfigServiceImpl) SetAppSetting(key string, value interface{}) error {
s.mu.Lock()
defer s.mu.Unlock()
s.settings[key] = value
return nil
}

// GetAllSettings retrieves all application settings
func (s *ConfigServiceImpl) GetAllSettings() (map[string]interface{}, error) {
s.mu.RLock()
defer s.mu.RUnlock()
settings := make(map[string]interface{})
for k, v := range s.settings {
settings[k] = v
}
return settings, nil
}

// RegisterService registers a service with the configuration
func (s *ConfigServiceImpl) RegisterService(name string, service interface{}) error {
s.mu.Lock()
defer s.mu.Unlock()
s.services[name] = service
return nil
}

// GetService retrieves a registered service
func (s *ConfigServiceImpl) GetService(name string) (interface{}, error) {
s.mu.RLock()
defer s.mu.RUnlock()
if service, ok := s.services[name]; ok {
return service, nil
}
return nil, nil
}

// Service getters
func (s *ConfigServiceImpl) GetUserService() UserService {
if service, err := s.GetService("user"); err == nil {
if userService, ok := service.(UserService); ok {
return userService
}
}
return nil
}

func (s *ConfigServiceImpl) GetGroupService() GroupService {
if service, err := s.GetService("group"); err == nil {
if groupService, ok := service.(GroupService); ok {
return groupService
}
}
return nil
}

func (s *ConfigServiceImpl) GetMessageService() MessageService {
if service, err := s.GetService("message"); err == nil {
if messageService, ok := service.(MessageService); ok {
return messageService
}
}
return nil
}

func (s *ConfigServiceImpl) GetPostService() PostService {
if service, err := s.GetService("post"); err == nil {
if postService, ok := service.(PostService); ok {
return postService
}
}
return nil
}
