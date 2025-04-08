package parser

import "regexp"

func IsSpecificPrefix(annotation string) bool {
	pattern := `^` + AnnotationsPrefix + "/"
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

func CheckAnnotations(ans map[string]string) error {
	return nil
}
