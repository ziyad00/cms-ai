// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"baliance.com/gooxml/color"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/presentation"
)

func main() {
	fmt.Println("🔬 DEBUG: Testing gooxml Visual Capabilities")

	// Create a simple presentation
	ppt := presentation.New()
	slide := ppt.AddSlide()

	// Test 1: Large red background text box
	fmt.Println("Creating large red background text box...")
	bgBox := slide.AddTextBox()
	bgProps := bgBox.Properties()
	bgProps.SetPosition(0, 0)
	bgProps.SetSize(
		measurement.Distance(10)*measurement.Inch,
		measurement.Distance(7.5)*measurement.Inch,
	)

	// Set bright red background
	redColor := color.FromHex("#FF0000")
	bgProps.SetSolidFill(redColor)
	bgProps.LineProperties().SetNoFill()

	// Add minimal content
	bgPara := bgBox.AddParagraph()
	bgRun := bgPara.AddRun()
	bgRun.SetText("BACKGROUND")
	bgRunProps := bgRun.Properties()
	bgRunProps.SetSize(measurement.Distance(1) * measurement.Point)
	// Make text white so it's visible on red
	whiteColor := color.FromHex("#FFFFFF")
	bgRunProps.SetSolidFill(whiteColor)

	// Test 2: Blue text box on top
	fmt.Println("Creating blue foreground text box...")
	fgBox := slide.AddTextBox()
	fgProps := fgBox.Properties()
	fgProps.SetPosition(
		measurement.Distance(2)*measurement.Inch,
		measurement.Distance(2)*measurement.Inch,
	)
	fgProps.SetSize(
		measurement.Distance(6)*measurement.Inch,
		measurement.Distance(3)*measurement.Inch,
	)

	// Set blue background
	blueColor := color.FromHex("#0000FF")
	fgProps.SetSolidFill(blueColor)
	fgProps.LineProperties().SetNoFill()

	// Add content
	fgPara := fgBox.AddParagraph()
	fgRun := fgPara.AddRun()
	fgRun.SetText("FOREGROUND TEXT")
	fgRunProps := fgRun.Properties()
	fgRunProps.SetSize(measurement.Distance(24) * measurement.Point)
	fgRunProps.SetSolidFill(whiteColor)

	// Save the test presentation
	outputPath := "test_outputs/debug_gooxml_visuals.pptx"
	if err := os.MkdirAll("test_outputs", 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	if err := ppt.SaveToFile(outputPath); err != nil {
		log.Fatalf("Error saving presentation: %v", err)
	}

	fmt.Printf("✅ Debug presentation saved: %s\n", outputPath)
	fmt.Println("🔍 Open this file to see what gooxml can actually render")
	fmt.Println("Expected: Red background with blue text box overlay")
}