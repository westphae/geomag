package polynomial

// Factorial calculates the factorial of the input integer.
// Doesn't handle negative numbers gracefully, up to user to not pass them.
// Handles up to n=20, beyond that it will overflow.
func Factorial(n int) int {
	r := 1
	for i := 2; i <= n; i++ {
		r *= i
	}
	return r
}

// FactorialRatio calculates n!/m!.
// Useful when dividing a large factorial by a smaller factorial, to fit
// inside an int64.
// Doesn't handle negative or large numbers gracefully, up to user to not pass them.
func FactorialRatio(n, m int) int {
	r := 1
	for i := m + 1; i <= n; i++ {
		r *= i
	}
	return r
}

// FactorialRatioFloat calculates n!/m! as a float64, to handle large numbers.
// Doesn't handle negative or large numbers gracefully, up to user to not pass them.
func FactorialRatioFloat(n, m int) float64 {
	r := 1.0
	for i := m + 1; i <= n; i++ {
		r *= float64(i)
	}
	return r
}

// Pow raises a float64 to the integer power n.
// Works for any n, positive, negative, or 0.
func Pow(x float64, n int) float64 {
	if n == 0 {
		return 1
	}
	if n < 0 {
		x = 1 / x
		n = -n
	}
	r := 1.0
	for range n {
		r *= x
	}
	return r
}
