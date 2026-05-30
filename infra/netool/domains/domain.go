package domains

// CutFromHead cut the domain from head to end
// Example:
// dn = "1.2.3.4.5"
// ans = [ "1.2.3.4.5", "2.3.4.5", "3.4.5", "4.5", "5" ]
func CutFromHead(dn string) []string {
	ans := []string{
		dn,
	}
	for i := 0; i < len(dn); i++ {
		if dn[i] == '.' && i < len(dn)-1 && len(dn[i+1:]) > 0 {
			ans = append(ans, dn[i+1:])
		}
	}
	return ans
}

// CutFromEnd cut the domain from end to head
// Example:
// dn = "1.2.3.4.5"
// ans = [ "1.2.3.4.5", "1.2.3.4", "1.2.3", "1.2", "1" ]
func CutFromEnd(dn string) []string {
	ans := []string{
		dn,
	}
	for i := len(dn) - 1; i >= 0; i-- {
		if dn[i] == '.' {
			ans = append(ans, dn[:i])
		}
	}
	return ans
}
