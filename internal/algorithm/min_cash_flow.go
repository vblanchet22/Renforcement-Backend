package algorithm

import "math"

const (
	debtSettlementThreshold = 0.01
	roundingPrecisionFactor = 100
)

// DebtEdge represents a simplified debt from one person to another
type DebtEdge struct {
	FromIndex int
	ToIndex   int
	Amount    float64
}

// MinCashFlow implements the min-cash-flow algorithm to simplify debts.
// It takes net balances for each person (positive = creditor, negative = debtor)
// and returns the minimum set of transactions needed to settle all debts.
func MinCashFlow(netBalances []float64) []DebtEdge {
	n := len(netBalances)
	if n <= 1 {
		return nil
	}

	// Create a copy to avoid modifying the original
	balances := make([]float64, n)
	copy(balances, netBalances)

	var result []DebtEdge
	minCashFlowRecursive(balances, &result)
	return result
}

func minCashFlowRecursive(balances []float64, result *[]DebtEdge) {
	// Find the person with the maximum credit (most owed to them)
	maxCreditIdx := findMaxIndex(balances)
	// Find the person with the maximum debit (owes the most)
	maxDebitIdx := findMinIndex(balances)

	// If both are effectively zero, we're done
	if math.Abs(balances[maxCreditIdx]) < debtSettlementThreshold && math.Abs(balances[maxDebitIdx]) < debtSettlementThreshold {
		return
	}

	// Find the minimum of the two absolute values
	// This is the amount that can be settled in this transaction
	minAmount := math.Min(-balances[maxDebitIdx], balances[maxCreditIdx])

	// Round to 2 decimal places
	minAmount = math.Round(minAmount*roundingPrecisionFactor) / roundingPrecisionFactor

	if minAmount < debtSettlementThreshold {
		return
	}

	// Update balances
	balances[maxCreditIdx] -= minAmount
	balances[maxDebitIdx] += minAmount

	// Add the transaction
	*result = append(*result, DebtEdge{
		FromIndex: maxDebitIdx,
		ToIndex:   maxCreditIdx,
		Amount:    minAmount,
	})

	// Recurse for remaining
	minCashFlowRecursive(balances, result)
}

// findMaxIndex returns the index of the maximum value
func findMaxIndex(arr []float64) int {
	maxIdx := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] > arr[maxIdx] {
			maxIdx = i
		}
	}
	return maxIdx
}

// findMinIndex returns the index of the minimum value
func findMinIndex(arr []float64) int {
	minIdx := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[minIdx] {
			minIdx = i
		}
	}
	return minIdx
}
