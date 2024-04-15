//go:build dev

package debug

func init() {
	if e, found := os.LookupEnv("ENVIRONMENT"); !found {
		panic("ENVIRONMENT variable not set")
	} else if strings.ToLower(e) == "production" {
		panic("debug build deployed to production")
	}
	enabled = true
}
