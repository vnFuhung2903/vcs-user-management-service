package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"github.com/vnFuhung2903/vcs-user-management-service/api"
	_ "github.com/vnFuhung2903/vcs-user-management-service/docs"
	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/infrastructures/databases"
	"github.com/vnFuhung2903/vcs-user-management-service/interfaces"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/env"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/repositories"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/services"
)

// @title VCS SMS API
// @version 1.0
// @description Container Management System API
// @host localhost:8085
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	env, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to retrieve env: %v", err)
	}

	logger, err := logger.LoadLogger(env.LoggerEnv)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	postgresDb, err := databases.ConnectPostgresDb(env.PostgresEnv)
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	postgresDb.AutoMigrate(&entities.User{}, &entities.UserScope{})

	sqlBytes, err := os.ReadFile("migration/init.sql")
	if err != nil {
		log.Fatalf("Error reading SQL file: %v", err)
	}
	execTrigger := postgresDb.Exec(string(sqlBytes))
	if execTrigger.Error != nil {
		log.Fatalf("Failed to execute trigger SQL: %v", execTrigger.Error)
	}

	redisRawClient := databases.NewRedisFactory(env.RedisEnv).ConnectRedis()
	redisClient := interfaces.NewRedisClient(redisRawClient)

	jwtMiddleware := middlewares.NewJWTMiddleware(env.AuthEnv)
	scopeRepository := repositories.NewScopeRepository(postgresDb)
	userRepository := repositories.NewUserRepository(postgresDb)

	scopeService := services.NewScopeService(scopeRepository, logger)
	userService := services.NewUserService(userRepository, redisClient, logger)
	scopeHandler := api.NewScopeHandler(scopeService, jwtMiddleware)
	userHandler := api.NewUserHandler(scopeService, userService, jwtMiddleware)

	r := gin.Default()
	scopeHandler.SetupRoutes(r)
	userHandler.SetupRoutes(r)
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":8085"); err != nil {
		log.Fatalf("Failed to run service: %v", err)
	} else {
		logger.Info("User management service is running on port 8085")
	}
}
