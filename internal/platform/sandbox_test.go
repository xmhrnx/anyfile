package platform

import "testing"

func TestSandboxed(t *testing.T) {
	for _, c := range []struct {
		goos string
		env  map[string]string
		want bool
	}{
		{"darwin", map[string]string{"HOME": "/Users/sam"}, false},
		{"darwin", map[string]string{"HOME": "/Users/sam/Library/Containers/com.xmhrnx.anyfile/Data"}, true},
		{"darwin", map[string]string{"APP_SANDBOX_CONTAINER_ID": "com.xmhrnx.anyfile", "HOME": "/Users/sam"}, true},
		{"linux", map[string]string{"HOME": "/home/sam/Library/Containers/x/Data"}, false},
		{"windows", map[string]string{"APP_SANDBOX_CONTAINER_ID": "x"}, false},
	} {
		if got := sandboxed(c.goos, func(k string) string { return c.env[k] }); got != c.want {
			t.Errorf("%s %v: got %v, want %v", c.goos, c.env, got, c.want)
		}
	}
}
