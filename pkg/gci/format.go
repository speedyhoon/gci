package gci

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/speedyhoon/gci/pkg/constants"
	importPkg "github.com/speedyhoon/gci/pkg/gci/imports"
	sectionsPkg "github.com/speedyhoon/gci/pkg/gci/sections"
	"github.com/speedyhoon/gci/pkg/gci/specificity"
)

// Formats the import section of a Go file
func FormatGoFile(input []byte, cfg GciConfiguration) ([]byte, error) {
	startIndex := bytes.Index(input, []byte(constants.ImportStartFlag))
	if startIndex < 0 {
		return nil, MissingImportStatementError
	}
	endIndexFromStart := bytes.Index(input[startIndex:], []byte(constants.ImportEndFlag))
	if endIndexFromStart < 0 {
		return nil, ImportStatementNotClosedError
	}
	endIndex := startIndex + endIndexFromStart

	unformattedImports := input[startIndex+len(constants.ImportStartFlag) : endIndex]
	formattedImports, err := formatImportBlock(unformattedImports, cfg)
	if err != nil {
		return nil, err
	}

	var output []byte
	output = append(output, input[:startIndex+len(constants.ImportStartFlag)]...)
	output = append(output, formattedImports...)
	output = append(output, input[endIndex+1:]...)
	return output, nil
}

// Takes unsorted imports as byte array and formats them according to the specified sections
func formatImportBlock(input []byte, cfg GciConfiguration) ([]byte, error) {
	//strings.ReplaceAll(input, "\r\n", linebreak)
	lines := strings.Split(string(input), constants.Linebreak)
	imports, err := parseToImportDefinitions(lines)
	if err != nil {
		return nil, fmt.Errorf("an error occured while trying to parse imports: %w", err)
	}
	if cfg.Debug {
		log.Println("Parsed imports in file:", imports)
	}
	// create mapping between sections and imports
	sectionMap := make(map[sectionsPkg.Section][]importPkg.ImportDef, len(cfg.Sections))
	// find matching section for every importSpec
	for _, i := range imports {
		// determine match specificity for every available section
		var bestSection sectionsPkg.Section
		var bestSectionSpecificity specificity.MatchSpecificity = specificity.MisMatch{}
		for _, section := range cfg.Sections {
			sectionSpecificity := section.MatchSpecificity(i)
			if sectionSpecificity.IsMoreSpecific(specificity.MisMatch{}) && sectionSpecificity.Equal(bestSectionSpecificity) {
				// specificity is identical
				return nil, EqualSpecificityMatchError{i, bestSection, section}
			}
			if sectionSpecificity.IsMoreSpecific(bestSectionSpecificity) {
				// better match found
				bestSectionSpecificity = sectionSpecificity
				bestSection = section
			}
		}
		if bestSection == nil {
			return nil, NoMatchingSectionForImportError{i}
		}
		if cfg.Debug {
			log.Printf("Matched import %s to section %s", i, bestSection)
		}
		sectionMap[bestSection] = append(sectionMap[bestSection], i)
	}
	// format every section to a str
	var sectionStrings []string
	for _, section := range cfg.Sections {
		sectionStr := section.Format(sectionMap[section], cfg.FormatterConfiguration)
		// prevent adding an empty section which would cause a separator to be inserted
		if sectionStr != "" {
			if cfg.Debug {
				log.Printf("Formatting section %s with imports: %v", section, sectionMap[section])
			}
			sectionStrings = append(sectionStrings, sectionStr)
		}
	}
	// format sectionSeparators
	var sectionSeparatorStr string
	for _, sectionSep := range cfg.SectionSeparators {
		sectionSeparatorStr += sectionSep.Format([]importPkg.ImportDef{}, cfg.FormatterConfiguration)
	}
	// generate output by joining the sections
	output := strings.Join(sectionStrings, sectionSeparatorStr)
	return []byte(output), nil
}
