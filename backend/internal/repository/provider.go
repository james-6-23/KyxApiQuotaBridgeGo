package repository

import (
    "github.com/google/wire"

    pgrepo "github.com/kyx-api-quota-bridge/backend/internal/repository/postgres"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

var ProviderSet = wire.NewSet(
    pgrepo.NewUserRepository,
    wire.Bind(new(UserRepository), new(*pgrepo.UserRepository)),
    pgrepo.NewClaimRepository,
    wire.Bind(new(ClaimRepository), new(*pgrepo.ClaimRepository)),
    pgrepo.NewDonateRepository,
    wire.Bind(new(DonateRepository), new(*pgrepo.DonateRepository)),
    pgrepo.NewDonatedKeyRepository,
    wire.Bind(new(DonatedKeyRepository), new(*pgrepo.DonatedKeyRepository)),
    pgrepo.NewAdminConfigRepository,
    wire.Bind(new(AdminConfigRepository), new(*pgrepo.AdminConfigRepository)),
)

func Provide(db *storage.Postgres) (
    UserRepository,
    ClaimRepository,
    DonateRepository,
    DonatedKeyRepository,
    AdminConfigRepository,
) {
    userRepo := pgrepo.NewUserRepository(db)
    claimRepo := pgrepo.NewClaimRepository(db)
    donateRepo := pgrepo.NewDonateRepository(db)
    keyRepo := pgrepo.NewDonatedKeyRepository(db)
    adminRepo := pgrepo.NewAdminConfigRepository(db)
    return userRepo, claimRepo, donateRepo, keyRepo, adminRepo
}
