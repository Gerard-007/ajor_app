package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContributionCycle string
type ContributionType string

const (
	CycleDaily   ContributionCycle = "daily"
	CycleWeekly  ContributionCycle = "weekly"
	CycleMonthly ContributionCycle = "monthly"
	CycleYearly  ContributionCycle = "yearly"
)

const (
	TypeDailySavings      ContributionType = "daily_savings"
	TypeGroupContribution ContributionType = "group_contribution"
)

type Contribution struct {
	ID                      primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name                    string               `json:"name" bson:"name"`
	Description             string               `json:"description" bson:"description"`
	Cycle                   ContributionCycle    `json:"cycle" bson:"cycle"`
	Amount                  float64              `json:"amount" bson:"amount"`
	CycleCount              int                  `json:"cycle_count" bson:"cycle_count"`
	CollectionDay           string               `json:"collection_day" bson:"collection_day"`
	CollectionDeadline      time.Time            `json:"collection_deadline" bson:"collection_deadline"`
	Type                    ContributionType     `json:"type" bson:"type"`
	PenaltyAmount           float64              `json:"penalty_amount" bson:"penalty_amount"`
	YetToCollectMembers     []primitive.ObjectID `json:"yet_to_collect_members" bson:"yet_to_collect_members"`
	AlreadyCollectedMembers []primitive.ObjectID `json:"already_collected_members" bson:"already_collected_members"`
	GroupAdmin              primitive.ObjectID   `json:"group_admin" bson:"group_admin"`
	AdminUsername           string               `json:"admin_username" bson:"admin_username"`
	MemberUsernames         map[primitive.ObjectID]string `json:"member_usernames" bson:"member_usernames"`
	WalletID                primitive.ObjectID   `json:"wallet_id" bson:"wallet_id"`
	InviteCode              string               `json:"invite_code" bson:"invite_code"`
	CreatedAt               time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at" bson:"updated_at"`
}