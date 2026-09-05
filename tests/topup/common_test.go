package topup_test

import (
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
)

type topupCardRepoAdapter struct {
	card_repo.CardQueryRepository
	card_repo.CardCommandRepository
}
