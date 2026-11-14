package integrations

import (
	"net/http"
)

// USSDWebhook handles Africa's Talking USSD callbacks.
// @Summary      USSD webhook
// @Description  Processes an Africa's Talking USSD callback and returns a plain-text CON or END response.
// @Tags         ussd
// @Accept       x-www-form-urlencoded
// @Produce      plain
// @Param        sessionId   formData  string  true  "USSD session identifier"
// @Param        serviceCode formData  string  true  "USSD service code"
// @Param        phoneNumber formData  string  true  "Subscriber phone number"
// @Param        text        formData  string  false "Cumulative USSD input"
// @Success      200         {string}  string  "USSD response prefixed with CON or END"
// @Failure      400         {string}  string  "Invalid form data"
// @Failure      405         {string}  string  "Method not allowed"
// @Failure      503         {string}  string  "USSD service unavailable"
// @Router       /api/v1/ussd [post]
func USSDWebhook(provider GuestManifestProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method != http.MethodPost {
			http.Error(w, "END Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "END Invalid form data", http.StatusBadRequest)
			return
		}

		response := HandleUSSD(r.Context(), USSDRequest{
			SessionID:   r.FormValue("sessionId"),
			ServiceCode: r.FormValue("serviceCode"),
			PhoneNumber: r.FormValue("phoneNumber"),
			Text:        r.FormValue("text"),
		}, provider)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}
}
