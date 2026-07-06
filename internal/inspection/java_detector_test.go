package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestJavaDetectorNotFoundWithoutMarkers(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewJavaDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without any Java build file")
	}
}

func TestJavaDetectorMavenSpringBoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pom.xml", `<project>
	<artifactId>my-api</artifactId>
	<version>1.0.0</version>
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
		</dependency>
	</dependencies>
</project>`)

	detection, ok := NewJavaDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.PackageManager != "maven" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "maven")
	}
	if detection.Name != "my-api" || detection.Version != "1.0.0" {
		t.Errorf("Name/Version = %q/%q, want %q/%q", detection.Name, detection.Version, "my-api", "1.0.0")
	}
	if detection.Framework != models.FrameworkSpring {
		t.Errorf("Framework = %q, want %q", detection.Framework, models.FrameworkSpring)
	}
	if detection.Suggested.Start != "mvn spring-boot:run" {
		t.Errorf("Suggested.Start = %q, want %q", detection.Suggested.Start, "mvn spring-boot:run")
	}
	if detection.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", detection.Confidence, ConfidenceHigh)
	}
}

func TestJavaDetectorGradle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "build.gradle", `plugins { id 'org.springframework.boot' version '3.2.0' }
dependencies { implementation 'org.springframework.boot:spring-boot-starter-web' }
`)

	detection, ok := NewJavaDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.PackageManager != "gradle" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "gradle")
	}
	if detection.Framework != models.FrameworkSpring {
		t.Errorf("Framework = %q, want %q", detection.Framework, models.FrameworkSpring)
	}
	if detection.Suggested.Start != "./gradlew bootRun" {
		t.Errorf("Suggested.Start = %q, want %q", detection.Suggested.Start, "./gradlew bootRun")
	}
	if !containsNote(detection.Notes, "not inferred from Gradle") {
		t.Errorf("expected a note about name/version not being inferred, got %v", detection.Notes)
	}
}

func TestJavaDetectorGradleKtsPreferredFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "build.gradle.kts", "plugins { }\n")

	detection, ok := NewJavaDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Framework != "" {
		t.Errorf("Framework = %q, want empty", detection.Framework)
	}
	if !containsNote(detection.Notes, "framework") {
		t.Errorf("expected a note about the undetermined framework, got %v", detection.Notes)
	}
}
