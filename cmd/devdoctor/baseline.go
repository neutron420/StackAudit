package main

import (
	"os"

	"devdoctor/internal/baseline"
	"devdoctor/internal/scanner"
)

func applyBaseline(report scanner.Report) (scanner.Report, error) {
	if cfg.UpdateBaseline {
		return report, baseline.Write(cfg.BaselinePath, cfg.RootPath, report)
	}

	base, err := baseline.Load(cfg.BaselinePath)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return report, err
	}

	return baseline.Filter(report, base, cfg.RootPath), nil
}
