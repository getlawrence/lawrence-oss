package opamp

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/open-telemetry/opamp-go/protobufs"
)

// SendToAgent sends a message to the connected agent
func (agent *Agent) SendToAgent(msg *protobufs.ServerToAgent) {
	_ = agent.conn.Send(context.Background(), msg)
}

// OfferConnectionSettings sends connection settings to the agent
func (agent *Agent) OfferConnectionSettings(offers *protobufs.ConnectionSettingsOffers) {
	agent.SendToAgent(
		&protobufs.ServerToAgent{
			ConnectionSettings: offers,
		},
	)
}

// addErrorResponse adds an error to the response message
func (agent *Agent) addErrorResponse(errMsg string, response *protobufs.ServerToAgent) {
	if response.ErrorResponse == nil {
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponseType_ServerErrorResponseType_BadRequest,
			ErrorMessage: errMsg,
			Details:      nil,
		}
	} else if response.ErrorResponse.Type == protobufs.ServerErrorResponseType_ServerErrorResponseType_BadRequest {
		// Append this error message to the existing error message.
		response.ErrorResponse.ErrorMessage += errMsg
	}
}

// processConnectionSettingsRequest processes a connection settings request from the agent
func (agent *Agent) processConnectionSettingsRequest(
	request *protobufs.OpAMPConnectionSettingsRequest, response *protobufs.ServerToAgent,
) {
	if request == nil || request.CertificateRequest == nil {
		return
	}

	csrDer, _ := pem.Decode(request.CertificateRequest.Csr)
	if csrDer == nil {
		agent.addErrorResponse("Failed to decode PEM certificate request", response)
		return
	}

	csr, err := x509.ParseCertificateRequest(csrDer.Bytes)
	if err != nil {
		agent.addErrorResponse("Failed to parse received certificate request: "+err.Error(), response)
		return
	}

	if csr.CheckSignature() != nil {
		agent.addErrorResponse("Certificate request signature check failed: "+err.Error(), response)
		return
	}

	// Verify the CSR's details and decide if we want to honor the request.
	// For OSS version, we accept all valid CSRs
	if csr.Subject.CommonName == "" {
		agent.addErrorResponse("CommonName is required in certificate request", response)
		return
	}

	// Create an offer for the agent.
	if response.ConnectionSettings == nil {
		response.ConnectionSettings = &protobufs.ConnectionSettingsOffers{}
	}
	response.ConnectionSettings.Opamp = &protobufs.OpAMPConnectionSettings{}
}

// calcConnectionSettings calculates connection settings for the agent
// For OSS version, we don't need special connection settings
func (agent *Agent) calcConnectionSettings(response *protobufs.ServerToAgent) {
	// In the SaaS version, this would set custom endpoints, certificates, etc.
	// For OSS, we don't need any special connection settings
}
