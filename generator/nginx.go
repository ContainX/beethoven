package generator

import (
	"bytes"
	"fmt"
	"github.com/ContainX/beethoven/tracker"
	"github.com/aymerick/raymond"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	tempTemplateName = ".nginx.conf.tmp-"
	nginxCommand     = "nginx"
)

// writeConfiguration writes a temporary nginx configuration based on
// the specified template.  It then validates the configuration for any
// errors
// return true if config has changed and been successfully updated
func (g *Generator) writeConfiguration() (bool, error) {
	tpl, err := raymond.ParseFile(g.cfg.Template)
	if err != nil {
		return false, fmt.Errorf("Error loading template: %s", err.Error())
	}
	result, err := tpl.Exec(g.templateData.Apps)
	if err != nil {
		return false, err
	}

	tplFilename, err := writeTempFile(result, filepath.Dir(g.cfg.Template), tempTemplateName)
	defer g.removeTempFile(tplFilename)

	if err != nil {
		return false, err
	}

	g.tracker.SetLastConfigRendered(time.Now())
	log.Info("wrote config: %s, contents: \n\n%s", tplFilename, result)

	if g.cfg.DryRun() {
		log.Debug("Has Changed from Config : %v", g.templateAndConfMatch(tplFilename))
		return false, nil
	}

	if err = g.validateConfig(tplFilename); err != nil {
		g.tracker.SetValidationError(&tracker.ValidationError{Error: err, FailedConfig: result})
		return false, err
	} else {
		g.tracker.ClearValidationError()
	}

	// At this points if the new/old configs don't match
	// issue a rename and nginx reload
	log.Debug("Temp Conf and Current Config Match : %v", g.templateAndConfMatch(tplFilename))
	if g.templateAndConfMatch(tplFilename) == false {
		log.Debug("Renaming %s to %s", tplFilename, g.cfg.NginxConfig)
		if err := os.Rename(tplFilename, g.cfg.NginxConfig); err != nil {
			return false, fmt.Errorf("Error renaming %s to %s: %s", tplFilename, g.cfg.NginxConfig, err.Error())
		}
		return true, nil
	}

	return false, nil
}

func (g *Generator) removeTempFile(file string) {
	os.Remove(file)
}

// Validates the temporary configuration file using NginX
func (g *Generator) validateConfig(tplFilename string) error {
	if err := g.execNginx("Validate Config:", "-c", tplFilename, "-t"); err != nil {
		return err
	}
	g.tracker.SetLastConfigValid(time.Now())

	return nil
}

func (g *Generator) reload() error {
	if err := g.execNginx("Reload NGINX:", "-s", "reload"); err != nil {
		return err
	}
	g.tracker.SetLastProxyReload(time.Now())
	return nil
}

func (g *Generator) execNginx(logPrefix string, args ...string) error {
	command := exec.Command(nginxCommand, args...)
	stderr := &bytes.Buffer{}
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("%s, %s, output: %s", logPrefix, err.Error(), stderr.String())
	}
	return nil
}

// Determines whether there are any differences between the newly generated
// template and the existing configuration.  If these are the same we bypass
// reloading NginX
// return bool - true if the two files match
func (g *Generator) templateAndConfMatch(tplFilename string) bool {
	tInfo, err := os.Stat(tplFilename)
	if err != nil {
		log.Warning(err.Error())
		return false
	}

	cInfo, err := os.Stat(g.cfg.NginxConfig)
	if err != nil {
		log.Warning(err.Error())
		return false
	}
	return (tInfo.Size() == cInfo.Size())
}

func writeTempFile(contents, baseDir, fileName string) (string, error) {
	tmpFile, err := ioutil.TempFile(baseDir, fileName)
	defer tmpFile.Close()

	if err != nil {
		return tmpFile.Name(), err
	}
	_, err = tmpFile.WriteString(contents)
	return tmpFile.Name(), err

}
