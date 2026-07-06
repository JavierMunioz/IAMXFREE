package inspection

import (
	"encoding/xml"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

type mavenProject struct {
	XMLName    xml.Name `xml:"project"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
}

type javaDetector struct{}

// NewJavaDetector returns a Detector for Java projects (Maven's pom.xml or
// Gradle's build.gradle/build.gradle.kts).
func NewJavaDetector() Detector {
	return &javaDetector{}
}

func (d *javaDetector) Name() string { return "java" }

func (d *javaDetector) Detect(in DetectionInput) (Detection, bool) {
	var matched []string
	for _, marker := range []string{"pom.xml", "build.gradle", "build.gradle.kts"} {
		if in.Has(marker) {
			matched = append(matched, marker)
		}
	}
	if len(matched) == 0 {
		return Detection{}, false
	}

	detection := Detection{
		Type:         ProjectTypeJava,
		Runtime:      models.RuntimeJava,
		MatchedFiles: matched,
	}

	switch {
	case in.Has("pom.xml"):
		d.detectMaven(in, &detection)
	default:
		d.detectGradle(in, &detection)
	}

	if detection.Framework == "" {
		detection.Notes = append(detection.Notes, "framework could not be determined")
	}

	return detection, true
}

func (d *javaDetector) detectMaven(in DetectionInput, detection *Detection) {
	detection.PackageManager = "maven"
	detection.Suggested.Install = "mvn install"
	detection.Suggested.Build = "mvn package"
	detection.Confidence = ConfidenceMedium

	data, err := in.ReadFile("pom.xml")
	if err != nil {
		detection.Notes = append(detection.Notes, "pom.xml could not be read")
		return
	}

	var project mavenProject
	if err := xml.Unmarshal(data, &project); err != nil {
		detection.Notes = append(detection.Notes, "pom.xml could not be parsed as XML")
		return
	}

	detection.Confidence = ConfidenceHigh
	detection.Name = project.ArtifactID
	detection.Version = project.Version
	if project.ArtifactID == "" {
		detection.Notes = append(detection.Notes, "pom.xml has no <artifactId>")
	}
	if project.Version == "" {
		detection.Notes = append(detection.Notes, "pom.xml has no <version>")
	}

	if strings.Contains(string(data), "spring-boot") {
		detection.Framework = models.FrameworkSpring
		detection.Suggested.Start = "mvn spring-boot:run"
	}
}

func (d *javaDetector) detectGradle(in DetectionInput, detection *Detection) {
	detection.PackageManager = "gradle"
	detection.Suggested.Build = "./gradlew build"
	detection.Confidence = ConfidenceMedium
	detection.Notes = append(detection.Notes, "project name/version are not inferred from Gradle build scripts")

	gradleFile := "build.gradle"
	if in.Has("build.gradle.kts") {
		gradleFile = "build.gradle.kts"
	}

	data, err := in.ReadFile(gradleFile)
	if err != nil {
		detection.Notes = append(detection.Notes, gradleFile+" could not be read")
		return
	}

	if strings.Contains(string(data), "spring-boot") {
		detection.Framework = models.FrameworkSpring
		detection.Suggested.Start = "./gradlew bootRun"
	}
}
