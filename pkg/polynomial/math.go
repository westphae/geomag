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
