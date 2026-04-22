package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ApprovalRefType identifies what kind of record this approval is for.
type ApprovalRefType = string

const (
	ApprovalRefPO       ApprovalRefType = "PO"
	ApprovalRefTransfer ApprovalRefType = "TRANSFER"
	ApprovalRefReject   ApprovalRefType = "REJECT"
)

// ApprovalStatus is the current state of the approval.
type ApprovalStatus = string

const (
	ApprovalPending  ApprovalStatus = "PENDING"
	ApprovalApproved ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
)

// ApprovalModel is a generic approval record linked to any approvable entity.
// Supports PO approval (required before GRN), Transfer approval (required before
// execution), and Reject Stock approval gates.
type ApprovalModel struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ApprovalId   string             `json:"approvalId" bson:"approvalId"`
	// ReferenceType: PO | TRANSFER | REJECT
	ReferenceType ApprovalRefType `json:"referenceType" bson:"referenceType"`
	// ReferenceId is the business ID (e.g. "PO-00001", "TRF-00001", "REJ-00001")
	ReferenceId  string `json:"referenceId" bson:"referenceId"`
	// Status: PENDING | APPROVED | REJECTED
	Status     ApprovalStatus `json:"status" bson:"status"`
	ApprovedBy string         `json:"approvedBy,omitempty" bson:"approvedBy,omitempty"`
	ApprovedAt time.Time      `json:"approvedAt,omitempty" bson:"approvedAt,omitempty"`
	Notes      string         `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt  time.Time      `json:"createdAt" bson:"createdAt"`
}

// ApprovalActionRequest is the request body for approve/reject actions.
type ApprovalActionRequest struct {
	Notes string `json:"notes"`
}

// SearchApprovalQuery filters for listing approvals.
type SearchApprovalQuery struct {
	ReferenceType string `json:"referenceType" query:"referenceType"`
	ReferenceId   string `json:"referenceId" query:"referenceId"`
	Status        string `json:"status" query:"status"`
	Page          int    `json:"page" query:"page"`
	Limit         int    `json:"limit" query:"limit"`
}
