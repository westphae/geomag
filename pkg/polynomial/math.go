package polynomial

// Factorial calculates the factorial of the input integer.
// Doesn't handle negative numbers gracefully, up to user to not pass them.
// Handles up to n=20, beyond that it will overflow.
func Factorial(n int) (z int) {
	if n>1 {
		return n*Factorial(n-1)
	}
	return 1
}

// FactorialRatio calculates the ratio of the factorial of the input integers.
// Useful when dividing a large factorial by a smaller factorial, to fit
// inside an int64.
// Doesn't handle negative or large numbers gracefully, up to user to not pass them.
func FactorialRatio(n, m int) (z int) {
	if n>m {
		return n*FactorialRatio(n-1, m)
	}
	return 1
}

// FactorialRatioFloat calculates the ratio of the factorial of the input integers
// and returns it as a float, to handle large numbers.
// Doesn't handle negative or large numbers gracefully, up to user to not pass them.
func FactorialRatioFloat(n, m int) (z float64) {
	if n>m {
		return float64(n)*FactorialRatioFloat(n-1, m)
	}
	return 1
}

// Pow raises a float64 to the integer power n.
// Works for any n, positive, negative or 0.
// Warning: Very inefficient for large n.
func Pow(x float64, n int) (y float64) {
	if n>0 {
		return x*Pow(x,n-1)
	}

	if n==0 {
		return 1
	}

	return 1/Pow(x,-n)
}
