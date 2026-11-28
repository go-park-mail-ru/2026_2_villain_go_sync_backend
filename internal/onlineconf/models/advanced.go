package models

import "time"

type Permission string

const (
	ReadKey        Permission = "read"
	CreateKey      Permission = "create"
	UpdateKey      Permission = "update"
	DeleteKey      Permission = "delete"
	ApproveRollout Permission = "approve"
)

type Grant struct {
	ID      int64      `json:"id"`
	UserID  *int64     `json:"user_id,omitempty"`
	GroupID *int64     `json:"group_id,omitempty"`
	ScopeID *int64     `json:"scope_id,omitempty"`
	KeyType *KeyType   `json:"key_type,omitempty"`
	KeyID   *int64     `json:"key_id,omitempty"`
	Action  Permission `json:"action"`
}

type Access struct {
	Full   bool    `json:"full"`
	Grants []Grant `json:"grants"`
}

type Group struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Revision  int64   `json:"revision"`
	MemberIDs []int64 `json:"member_ids"`
}

type RolloutStep struct {
	Percentage  int16 `json:"percentage"`
	WaitSeconds int32 `json:"wait_seconds"`
}

type RolloutPolicy struct {
	ID    int64         `json:"id"`
	Name  string        `json:"name"`
	Steps []RolloutStep `json:"steps"`
}

type RolloutStatus string

const (
	RolloutRunning   RolloutStatus = "running"
	RolloutCompleted RolloutStatus = "completed"
	RolloutCanceled  RolloutStatus = "canceled"
)

type Rollout struct {
	KeyID      int64         `json:"key_id"`
	RunID      int64         `json:"run_id"`
	PolicyName string        `json:"policy_name"`
	Steps      []RolloutStep `json:"steps"`
	Status     RolloutStatus `json:"status"`
	NextStep   int           `json:"next_step"`
	DueAt      *time.Time    `json:"due_at"`
	ApprovedBy *int64        `json:"approved_by"`
}
