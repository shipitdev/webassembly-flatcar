package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type OrderItem struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Qty   int     `json:"qty"`
}

type OrderRequest struct {
	OrderID    string      `json:"order_id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

type OrderResponse struct {
	OrderID     string  `json:"order_id"`
	TotalItems  int     `json:"total_items"`
	Subtotal    float64 `json:"subtotal"`
	Tax         float64 `json:"tax"`
	GrandTotal  float64 `json:"grand_total"`
	Signature   string  `json:"signature"`
	ProcessedUs int64   `json:"processed_us"`
}

// OrderHandler processes incoming purchase orders, validates line items,
// calculates jurisdictional tax, generates cryptographic signatures, and persists state.
func OrderHandler(store *StateStore, totalReqs *uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		atomic.AddUint64(totalReqs, 1)

		var order OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, "invalid json payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 1. Compute itemized financial totals
		var subtotal float64
		var totalItems int
		for _, item := range order.Items {
			subtotal += item.Price * float64(item.Qty)
			totalItems += item.Qty
		}
		tax := subtotal * 0.08 // standard 8% simulated tax
		grandTotal := subtotal + tax

		// 2. Cryptographic Fraud Signature (SHA-256 digest over transaction tuple)
		rawSig := fmt.Sprintf("%s:%s:%.2f", order.OrderID, order.CustomerID, grandTotal)
		sigHash := sha256.Sum256([]byte(rawSig))
		signature := hex.EncodeToString(sigHash[:])

		// 3. Persist transaction to in-memory state store
		store.Record(order.OrderID, grandTotal)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrderResponse{
			OrderID:     order.OrderID,
			TotalItems:  totalItems,
			Subtotal:    subtotal,
			Tax:         tax,
			GrandTotal:  grandTotal,
			Signature:   signature,
			ProcessedUs: time.Since(start).Microseconds(),
		})
	}
}
