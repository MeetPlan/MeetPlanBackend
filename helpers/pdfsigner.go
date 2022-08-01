package helpers

import (
	"io/ioutil"
	"janouch.name/pdf-simple-sign/pdf"
)

func Sign(doc []byte, outputPath string, pkcs12FilePath string, pkcs12Password string) error {
	p12, err := ioutil.ReadFile(pkcs12FilePath)
	if err != nil {
		return err
	}
	key, certs, err := pdf.PKCS12Parse(p12, pkcs12Password)
	if err != nil {
		return err
	}
	if doc, err = pdf.Sign(doc, key, certs, 4096); err != nil {
		return err
	}
	if err = ioutil.WriteFile(outputPath, doc, 0666); err != nil {
		return err
	}
	return nil
}
