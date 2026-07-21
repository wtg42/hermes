package sendmail

import (
	"fmt"
	"strconv"
	"strings"
)

const safeSingleRecipientDomain = "rd01.softnext.com.tw"

var safeSingleSenders = []string{
	"weitingshih@rd01.softnext.com.tw",
	"jllee@rd01.softnext.com.tw",
	"adam@rd01.softnext.com.tw",
}

// SingleSendSafetyInput contains the final envelope values assessed before one send.
type SingleSendSafetyInput struct {
	From string
	To   []string
	CC   []string
	BCC  []string
	Port string
}

// SingleSendSafetyAssessment lists whitelist exceptions requiring authorization.
type SingleSendSafetyAssessment struct {
	Reasons []string
}

// RequiresConfirmation reports whether the send crosses a safe whitelist boundary.
func (a SingleSendSafetyAssessment) RequiresConfirmation() bool {
	return len(a.Reasons) > 0
}

// AssessSingleSendSafety validates one envelope and returns every whitelist exception.
func AssessSingleSendSafety(input SingleSendSafetyInput) (SingleSendSafetyAssessment, error) {
	from, err := validateMailbox("from", input.From)
	if err != nil {
		return SingleSendSafetyAssessment{}, err
	}
	if len(input.To) == 0 {
		return SingleSendSafetyAssessment{}, fmt.Errorf("at least one to email address is required")
	}
	to, err := validateMailboxList("to", input.To)
	if err != nil {
		return SingleSendSafetyAssessment{}, err
	}
	cc, err := validateMailboxList("cc", input.CC)
	if err != nil {
		return SingleSendSafetyAssessment{}, err
	}
	bcc, err := validateMailboxList("bcc", input.BCC)
	if err != nil {
		return SingleSendSafetyAssessment{}, err
	}

	port, err := strconv.Atoi(input.Port)
	if err != nil || port < 1 || port > 65535 {
		return SingleSendSafetyAssessment{}, fmt.Errorf("port must be an integer between 1 and 65535")
	}

	var reasons []string
	if !containsFold(safeSingleSenders, from) {
		reasons = append(reasons, fmt.Sprintf("sender %s is outside the sender whitelist", from))
	}
	for _, recipient := range append(append(append([]string{}, to...), cc...), bcc...) {
		if recipientDomain(recipient) != safeSingleRecipientDomain {
			reasons = append(reasons, fmt.Sprintf("recipient %s is outside %s", recipient, safeSingleRecipientDomain))
		}
	}
	if input.Port != "25" {
		reasons = append(reasons, fmt.Sprintf("port %s is outside the safe port 25", input.Port))
	}

	return SingleSendSafetyAssessment{Reasons: reasons}, nil
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func recipientDomain(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(address[at+1:])
}
