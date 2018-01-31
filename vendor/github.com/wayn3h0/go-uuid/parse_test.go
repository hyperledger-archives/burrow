package uuid

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	uuids := []string{
		"00000000-0000-0000-0000-000000000000", // nil
		"945f6800-b463-11e4-854a-0002a5d5c51b", // v1: time based
		"000001f5-b465-21e4-8ffe-98fe945016ea", // v2: dce security
		"6fa459ea-ee8a-3ca4-894e-db77e160355e", // v3: name based md5
		"0f8fad5b-d9cb-469f-a165-70867728950e", // v4: random
		"886313e1-3b8a-5372-9b90-0c9aee199e5d", // v5: name based sha1
	}
	for i, v := range uuids {
		u, err := Parse(v)
		if err != nil {
			t.Fatalf("%s: %s", v, err)
		}

		switch i {
		case 0:
			if !u.Equal(Nil) {
				t.Fatalf("UUID[%s] should equal to nil uuid", v)
			}
		case 1:
			if u.Version() != VersionTimeBased {
				t.Fatalf("version of UUID[%s] should be Time Based (v1)", v)
			}
		case 2:
			if u.Version() != VersionDCESecurity {
				t.Fatalf("version of UUID[%s] should be DCE Security (v2)", v)
			}
		case 3:
			if u.Version() != VersionNameBasedMD5 {
				t.Fatalf("version of UUID[%s] should be Name Based MD5 (v3)", v)
			}
		case 4:
			if u.Version() != VersionRandom {
				t.Fatalf("version of UUID[%s] should be Random (v4)", v)
			}
		case 5:
			if u.Version() != VersionNameBasedSHA1 {
				t.Fatalf("version of UUID[%s] should be Name Based SHA1 (v5)", v)
			}
		}

		fmt.Printf("%s, %s, %s\n", u.String(), u.Version(), u.Layout())
	}
}

func TestErrors(t *testing.T) {
	uuids := []string{
		"00000000-00000000-0000-0000-00000000", "is an incorrect pattern",
		"945f6800-b463-11e4-854a-0002a5d5c51b2a", "is too long",
		"000001f5-b465-21h4-8ffe-98fe945016ea", "contains an invalid character",
		"0f8fad5bd9cb469fa1657086772895", "is too short",
	}
	for i := 0; i < len(uuids); i += 2 {
		if _, err := Parse(uuids[i]); err == nil {
			t.Fatalf("Failed to detect that \"%s\" %s", uuids[i], uuids[i+1])
		} else {
			fmt.Printf("Correctly detected that \"%s\" %s\n    %s\n", uuids[i], uuids[i+1], err.Error())
		}
	}
}
