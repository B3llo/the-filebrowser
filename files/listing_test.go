package files

import (
	"reflect"
	"testing"
)

func TestNameSortMatchesDirectionAndKeepsDirectoriesFirst(t *testing.T) {
	for _, tc := range []struct {
		name string
		asc  bool
		want []string
	}{
		{"ascending", true, []string{"folder", "a.mp4", "B.mp4", "video2.mp4", "video10.mp4"}},
		{"descending", false, []string{"folder", "video10.mp4", "video2.mp4", "B.mp4", "a.mp4"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listing := Listing{
				Sorting: Sorting{By: "name", Asc: tc.asc},
				Items: []*FileInfo{
					{Name: "video10.mp4"}, {Name: "a.mp4"},
					{Name: "folder", IsDir: true}, {Name: "B.mp4"}, {Name: "video2.mp4"},
				},
			}
			listing.ApplySort()
			var got []string
			for _, item := range listing.Items {
				got = append(got, item.Name)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("names = %v, want %v", got, tc.want)
			}
		})
	}
}
