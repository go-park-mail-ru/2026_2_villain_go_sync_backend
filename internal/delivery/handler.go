package delivery

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"

type Handler struct {
	Auth    *usecase.AuthUsecase
	Vacancy *usecase.VacancyUsecase
}

func NewHandler(auth *usecase.AuthUsecase, vacancy *usecase.VacancyUsecase) *Handler {
	return &Handler{Auth: auth, Vacancy: vacancy}
}
