package template

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TemplateData holds the data to be substituted in templates
type TemplateData struct {
	LocalIP     string
	LocalPort   int
	SMBServer   string
	SessionUSN  string
	RedirectURL string
}

// Manager handles template loading and processing
type Manager struct {
	templateDir string
	data        TemplateData
}

// NewManager creates a new template manager
func NewManager(templateDir string, data TemplateData) *Manager {
	return &Manager{
		templateDir: templateDir,
		data:        data,
	}
}

// BuildDeviceXML builds the device descriptor XML file
func (m *Manager) BuildDeviceXML() (string, error) {
	return m.processTemplate("device.xml")
}

// BuildServiceXML builds the service descriptor XML file
func (m *Manager) BuildServiceXML() (string, error) {
	servicePath := filepath.Join(m.templateDir, "service.xml")
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		// Return minimal XML if service.xml doesn't exist
		return ".", nil
	}
	return m.processTemplate("service.xml")
}

// BuildPhishHTML builds the phishing page HTML
func (m *Manager) BuildPhishHTML() (string, error) {
	content, err := m.processTemplate("present.html")
	if err != nil {
		return "", err
	}
	
	// Wrap the content in proper HTML structure if it doesn't already have it
	if !strings.Contains(strings.ToLower(content), "<html") {
		content = "<html>\n" + content + "\n</html>"
	}
	
	return content, nil
}

// BuildExfilDTD builds the DTD file for XXE exfiltration
func (m *Manager) BuildExfilDTD() (string, error) {
	if !strings.Contains(m.templateDir, "xxe-exfil") {
		return ".", nil
	}
	return m.processTemplate("data.dtd")
}

// processTemplate loads and processes a template file
func (m *Manager) processTemplate(filename string) (string, error) {
	templatePath := filepath.Join(m.templateDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}
	
	// Read the template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	
	// Convert Python-style template variables to Go template syntax
	templateContent := m.convertTemplateVars(string(content))
	
	// Create and parse the template
	tmpl, err := template.New(filename).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", filename, err)
	}
	
	// Execute the template with data
	var result strings.Builder
	if err := tmpl.Execute(&result, m.data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", filename, err)
	}
	
	return result.String(), nil
}

// convertTemplateVars converts Python string.Template variables to Go template syntax
func (m *Manager) convertTemplateVars(content string) string {
	// Convert Python template variables to Go template variables
	// $SMB_SERVER -> {{.SMBServer}}
	// $local_ip -> {{.LocalIP}}
	// $local_port -> {{.LocalPort}}
	// $session_usn -> {{.SessionUSN}}
	// $redirect_url -> {{.RedirectURL}}
	// $smb_server -> {{.SMBServer}}
	
	replacements := map[string]string{
		"$SMB_SERVER":   "{{.SMBServer}}",
		"$smb_server":   "{{.SMBServer}}",
		"$local_ip":     "{{.LocalIP}}",
		"$local_port":   "{{.LocalPort}}",
		"$session_usn":  "{{.SessionUSN}}",
		"$redirect_url": "{{.RedirectURL}}",
	}
	
	result := content
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	// Handle $$ -> $ conversion (Python template escaping)
	result = strings.ReplaceAll(result, "$$", "$")
	
	return result
}

// ValidateTemplateDir checks if the template directory exists and has required files
func ValidateTemplateDir(templateDir string) error {
	// Check if directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory does not exist: %s", templateDir)
	}
	
	// Check for required files
	requiredFiles := []string{"device.xml", "present.html"}
	
	for _, file := range requiredFiles {
		filePath := filepath.Join(templateDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required template file not found: %s", filePath)
		}
	}
	
	return nil
}

// ListTemplates returns a list of available templates
func ListTemplates(templatesBaseDir string) ([]string, error) {
	var templates []string
	
	err := filepath.WalkDir(templatesBaseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() && path != templatesBaseDir {
			// Check if this directory has the required template files
			if err := ValidateTemplateDir(path); err == nil {
				relPath, _ := filepath.Rel(templatesBaseDir, path)
				templates = append(templates, relPath)
			}
		}
		
		return nil
	})
	
	return templates, err
}