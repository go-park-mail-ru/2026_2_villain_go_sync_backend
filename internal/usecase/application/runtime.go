package application

type RuntimeSettings interface {
	ApplicationsEnabled(seekerID int64) bool
}
