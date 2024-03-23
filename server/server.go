package server

import (
	"context"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"test/handler"
	"test/middlewares"

	"time"
)

type Server struct {
	*mux.Router
	server *http.Server
}

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

func SetupRoutes() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "My name is Deepak Singh") }).Methods(http.MethodGet)
	router.HandleFunc("/register",handler.RegisterUser).Methods(http.MethodPost)
	router.HandleFunc("/login",handler.Login).Methods(http.MethodPost)
	router.HandleFunc("/products",handler.ShowProduct).Methods(http.MethodGet)
	router.HandleFunc("/type-product",handler.ShowProductByType).Methods(http.MethodGet)
	userRouter := router.PathPrefix("/home").Subrouter()
	userRouter.Use(middlewares.JWTMiddleware)
	userRouter.HandleFunc("/address",handler.CreateAddress).Methods(http.MethodPost)
	userRouter.HandleFunc("/address",handler.ShowAddress).Methods(http.MethodGet)
	userRouter.HandleFunc("/address",handler.DeleteAddressById).Methods(http.MethodDelete)
	userRouter.HandleFunc("/cart-item",handler.AddCartItem).Methods(http.MethodPost)
	userRouter.HandleFunc("/cart-item",handler.DeleteCartItem).Methods(http.MethodDelete)
	userRouter.HandleFunc("/cart-item",handler.ShowCartItems).Methods(http.MethodGet)

	adminRouter := userRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.AdminMiddleWare)
	adminRouter.HandleFunc("/user",handler.AddUser).Methods(http.MethodPost)
	adminRouter.HandleFunc("/user",handler.ShowUser).Methods(http.MethodGet)
	adminRouter.HandleFunc("/user",handler.DeleteUserById).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/product",handler.CreateProduct).Methods(http.MethodPost)
	adminRouter.HandleFunc("/upload",handler.UploadHandlerAWS).Methods(http.MethodPost)
	adminRouter.HandleFunc("/product",handler.DeleteProductById).Methods(http.MethodDelete)




	return &Server{
		Router: router,
	}
}

func (svc *Server) Run(port string) error {
	svc.server = &http.Server{
		Addr:              port,
		Handler:           svc.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return svc.server.ListenAndServe()
}

func (svc *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svc.server.Shutdown(ctx)
}
