package http

import (
	"context"
	"fmt"
	"net/http"
	"os"

	//"net/smtp"
	"time"

	//"test123/config"
	"test123/handler"
	kafka "test123/kafka/producers"
	middlewares "test123/middleware"
	"test123/service"
	"test123/utils/jwt"

	// "test123/utils/jwt"

	"test123/repositories"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/willf/bloom"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	DBStatus    string
	UserService *service.UserService

	RedisClient *redis.Client

	AuthService       *service.AuthService
	AuthorizseService *service.AuthorizeService
	JWT               *jwt.Jwt

	KafkaProducer   *kafka.KafkaNotificationProducer
	RoleService     *service.RoleService
	UserRoleService *service.UserRoleService
	BloomFilter     *bloom.BloomFilter
}

// Constructor
func NewServer(dbStatus string, db *pgxpool.Pool, rdb *redis.Client, kafka *kafka.KafkaNotificationProducer, bloom *bloom.BloomFilter) *Server {
	userRepo := repositories.NewUserRepo(db)

	roleRepo := repositories.NewRoleRepo(db)
	userroleRepo := repositories.NewUserRoleRepo(db)
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	j := jwt.NewJwt(jwtSecret)

	userService := service.NewUserService(userRepo, kafka, userroleRepo, rdb, bloom)

	authService := service.NewAuthService(userService, rdb, j, kafka)
	roleService := service.NewRoleService(roleRepo)
	authorizeService := service.NewAuthorizeService(db, rdb)
	userroleService := service.NewUserRoleService(userroleRepo)

	return &Server{
		DBStatus:    dbStatus,
		UserService: userService,

		RedisClient:       rdb,
		AuthService:       authService,
		JWT:               j,
		KafkaProducer:     kafka,
		RoleService:       roleService,
		AuthorizseService: authorizeService,
		UserRoleService:   userroleService,
		BloomFilter:       bloom,
	}
}

// Listen & Serve
func (s *Server) Listen(ctx context.Context, addr string) error {

	// Create handlers (Dependency Injection)
	userHandler := handler.NewUserHandler(s.UserService)
	authHandler := handler.NewAuthHandler(s.AuthService)
	adminHandler := handler.NewAdminHandler(s.RoleService, s.UserRoleService, s.UserService)

	r := chi.NewRouter()

	// // Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API group
	r.Route("/api/v1/", func(r chi.Router) {

		r.Route("/users", func(r chi.Router) {
			r.Get("/{Id}", userHandler.GetUserById)
			r.Post("/", userHandler.CreateUser)
			r.Post("/{username}/check", userHandler.CheckUsernameHandler)
			r.With(middlewares.AuthMiddleware(s.AuthService), middlewares.RequirePermission(s.AuthorizseService, "user.read.all")).Get("/", userHandler.GetAllUsers)
			r.With(middlewares.AuthMiddleware(s.AuthService), middlewares.RequirePermission(s.AuthorizseService, "user.self.update")).Put("/{id}", userHandler.UpdateUser)
			r.With(middlewares.AuthMiddleware(s.AuthService), middlewares.RequirePermission(s.AuthorizseService, "user.self.delete")).Delete("/{id}", userHandler.DeleteUser)
		})
		// Auth routes under /api/test/auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/generate-token", authHandler.GenerateResetToken)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.WipeOutSession)
			r.Post("/access-token", authHandler.GenerateAccessToken)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(s.AuthService), middlewares.RequirePermission(s.AuthorizseService, "role.assign"))
			r.Post("/", adminHandler.CreateRole)
			r.Post("/assign-role", adminHandler.AddRoleToUser)
			r.Delete("/user/{Id}", adminHandler.DeleteUser)
			r.Get("/user/{Id}", adminHandler.GetUserByID)
		})

	})

	// HTTP Server
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	errCh := make(chan error, 1)

	go func() {
		fmt.Printf(" Server running on %s\n", addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		fmt.Println("\n Shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
