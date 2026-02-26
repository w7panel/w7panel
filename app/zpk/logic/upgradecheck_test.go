package logic

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestCheckX(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	// sdk := k8s.NewClient
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwczovL2NvbnNvbGUudzcuY2MvYXBpL3RoaXJkcGFydHktY2QvazhzLW9mZmxpbmUvb3BlbmlkLXRvLWNkLXRva2VuIiwiaWF0IjoxNzY5NzQxNTkyLCJleHAiOjE3Njk3NDY5OTIsIm5iZiI6MTc2OTc0MTU5MiwianRpIjoiM3l1N2FCdUVCTWZJVnFSbSIsInN1YiI6IjM2MTQ4OSIsInBydiI6ImYwZDE5MGY3YWYyZTg3Y2U2ZjAxNmFhODIwNTBmYzcwZjIzZjU5NWMiLCJvcGVuX2lkIjoiQ2k0aGpaYnJaRkxxS1dBVEhVYjVydyIsImZvdW5kZXJfb3BlbmlkIjoiQ2k0aGpaYnJaRkxxS1dBVEhVYjVydyIsInJvbGVfaWRlbnRpZnkiOiJmb3VuZGVyIiwib3JpZ2luX2FwcGlkIjoiMzI4NDk0Iiwibmlja25hbWUiOiJ0bXAiLCJ1c2VyX2lkIjozNjE0ODksImlzX3ZhbGlkIjp0cnVlfQ.m8HgOOXTUyi44yuFMSkRr8-L6QC7HmnnGIuYIFfceSA"
	upgradeCheck := NewUpgradeCheck(sdk)
	upgradeCheck.WithCDToken(token)
	upgradeCheck.Check("default", "w7-pros-28693-gte7p7a84d")
	// t.print(result)

}
