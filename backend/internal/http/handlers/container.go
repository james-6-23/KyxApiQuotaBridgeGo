package handlers

import (
    "github.com/go-chi/chi/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/service"
)

type Container struct {
    Auth    *AuthHandler
    User    *UserHandler
    Claim   *ClaimHandler
    Donate  *DonateHandler
    Admin   *AdminHandler
}

func Provide(services *service.Services) (*Container, error) {
    return &Container{
        Auth:   NewAuthHandler(services),
        User:   NewUserHandler(services),
        Claim:  NewClaimHandler(services),
        Donate: NewDonateHandler(services),
        Admin:  NewAdminHandler(services),
    }, nil
}

func (c *Container) RegisterRoutes(r chi.Router) {
    r.Route("/api", func(api chi.Router) {
        api.Group(func(auth chi.Router) {
            auth.Post("/auth/bind", c.Auth.Bind)
            auth.Post("/auth/logout", c.Auth.Logout)
        })

        api.Group(func(user chi.Router) {
            user.Use(c.Auth.UserSessionMiddleware)
            user.Get("/user/quota", c.User.Quota)
            user.Get("/user/records/claim", c.User.Claims)
            user.Get("/user/records/donate", c.User.Donations)
            user.Post("/claim/daily", c.Claim.Daily)
            user.Post("/donate/validate", c.Donate.Validate)
        })

        api.Route("/admin", func(admin chi.Router) {
            admin.Group(func(public chi.Router) {
                public.Post("/login", c.Admin.Login)
            })

            admin.Group(func(priv chi.Router) {
                priv.Use(c.Auth.AdminSessionMiddleware)
                priv.Get("/config", c.Admin.Config)
                priv.Put("/config/quota", c.Admin.UpdateQuota)
                priv.Put("/config/session", c.Admin.UpdateSession)
                priv.Put("/config/new-api-user", c.Admin.UpdateNewAPIUser)
                priv.Put("/config/keys-api-url", c.Admin.UpdateKeysAPIURL)
                priv.Put("/config/keys-authorization", c.Admin.UpdateKeysAuthorization)
                priv.Put("/config/group-id", c.Admin.UpdateGroupID)

                priv.Get("/records/claim", c.Admin.ClaimRecords)
                priv.Get("/records/donate", c.Admin.DonateRecords)
                priv.Get("/users", c.Admin.Users)
                priv.Get("/keys/export", c.Admin.ExportKeys)
                priv.Post("/keys/delete", c.Admin.DeleteKeys)
                priv.Post("/keys/test", c.Admin.TestKeys)
                priv.Post("/retry-push", c.Admin.RetryPush)
                priv.Post("/rebind-user", c.Admin.RebindUser)
                priv.Get("/export/users", c.Admin.ExportUsers)
                priv.Post("/import/users", c.Admin.ImportUsers)
            })
        })
    })
}
