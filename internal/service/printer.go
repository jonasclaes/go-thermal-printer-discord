package service

import (
	"errors"
	"log"
	"net/url"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/jonasclaes/go-thermal-printer-discord/internal/swagger/client/printer"
	"github.com/jonasclaes/go-thermal-printer-discord/internal/swagger/models"
	"github.com/spf13/viper"
)

type PrinterService struct {
	apiClient  printer.ClientService
	apiKeyAuth runtime.ClientAuthInfoWriter
}

func NewPrinterService() (*PrinterService, error) {
	printerEndpoint, err := url.Parse(viper.GetString("printer.endpoint"))
	if err != nil {
		log.Fatalf("error parsing printer host: %v", err)
		return nil, err
	}

	transport := httptransport.New(printerEndpoint.Host, "", []string{printerEndpoint.Scheme})
	apiKeyAuth := httptransport.APIKeyAuth("x-api-key", "header", viper.GetString("printer.api_key"))
	apiClient := printer.New(transport, strfmt.Default)

	return &PrinterService{
		apiClient:  apiClient,
		apiKeyAuth: apiKeyAuth,
	}, nil
}

func (s *PrinterService) PrintTodo(title string) error {
	templateFile := "templates/todo.tmpl"
	variables := map[string]interface{}{
		"title": title,
	}

	requestData := printer.NewPostAPIV1PrinterPrintTemplateParams().WithRequest(&models.PrinterPrintTemplateDto{
		TemplateFile: &templateFile,
		Variables:    variables,
	})

	response, err := s.apiClient.PostAPIV1PrinterPrintTemplate(requestData, s.apiKeyAuth)
	if err != nil {
		log.Printf("error sending print job: %v", err)
		return err
	}

	if !response.IsSuccess() {
		log.Printf("error response from printer service: %v", response)
		return errors.New(response.Error())
	}

	return nil
}
