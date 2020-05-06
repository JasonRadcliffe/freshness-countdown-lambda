package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	dishDomain "github.com/jasonradcliffe/freshness-countdown-api/domain/dish"

	alexa "github.com/ericdaugherty/alexa-skills-kit-golang"
)

var (
	a = &alexa.Alexa{
		ApplicationID:   "amzn1.ask.skill.bfe622c0-a8b3-4a8b-afcb-072515a3f935",
		RequestHandler:  &ProcessEvents{},
		IgnoreTimestamp: true,
	}
	baseURL = "https://fcapi.jasonradcliffe.com"
)

const cardTitle = "HelloJason"

//ProcessEvents impliments the interface from the library.
type ProcessEvents struct{}

//FCAPIResponse is the structure the responses from the API will take.
type FCAPIResponse struct {
	Message []byte `json:"message"`
}

//OnSessionStarted is a func.
func (p *ProcessEvents) OnSessionStarted(c context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	log.Printf("OnSessionStarted requestID=%s, sessionID=%s", request.RequestID, session.SessionID)
	return nil
}

//OnLaunch is a func.
func (p *ProcessEvents) OnLaunch(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	speechText := "Welcome to the Freshness Countdown Alexa skill, you can say hello"

	log.Printf("OnLaunch requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	response.SetSimpleCard(cardTitle, speechText)
	response.SetOutputText(speechText)
	response.SetRepromptText(speechText)

	response.ShouldSessionEnd = true

	return nil
}

//OnIntent is a func.
func (p *ProcessEvents) OnIntent(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	var speechText string
	var userAccessToken string
	var userAlexaID string
	apiRequestMap := make(map[string]string)

	log.Printf("OnIntent requestId=%s, sessionId=%s, intent=%s", request.RequestID, session.SessionID, request.Intent.Name)

	userAccessToken = session.User.AccessToken
	userAlexaID = session.User.UserID

	apiRequestMap["accessToken"] = userAccessToken
	apiRequestMap["alexaUserID"] = userAlexaID

	switch request.Intent.Name {
	case "GetDishes":

		apiRequestMap["fcapiRequestType"] = "GET"

		requestBody, _ := json.Marshal(apiRequestMap)

		resp, err := http.Post("https://fcapi.jasonradcliffe.com/dishes", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			return errors.New("Unsuccessful")
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("Unsuccessful on body read")
		}

		var apiResp FCAPIResponse
		//var apiResp2 FCAPIResponse
		log.Printf("got this response from API: \n%s", string(body))

		jsonErr := json.Unmarshal(body, &apiResp)
		if jsonErr != nil {
			return errors.New("Unsuccessful on json unmarshal")
		}

		//var dishList dishDomain.Dishes
		log.Printf("Here is the apiResp.Message:%s", apiResp.Message)
		//log.Printf("Here is the apiResp.Message.Message (2nd level!!!):%s", apiResp2.Message)
		var dishes dishDomain.Dishes
		disherr := json.Unmarshal(apiResp.Message, &dishes)
		if disherr != nil {
			return errors.New("Unsuccessful on dishlist")
		}

		speechText = "DId it work? Dish #1:" + dishes[0].Title

		response.SetSimpleCard(cardTitle, speechText)
		response.SetOutputText(speechText)

	case "GetDish":
		dishID := request.Intent.Slots["dishID"].Value

		apiRequestMap["fcapiRequestType"] = "GET"

		requestBody, _ := json.Marshal(apiRequestMap)

		resp, err := http.Post("https://fcapi.jasonradcliffe.com/dishes/"+dishID, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			return errors.New("Unsuccessful")
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("Unsuccessful on body read")
		}

		var apiResp FCAPIResponse
		//var apiResp2 FCAPIResponse

		log.Printf("got this response from API: \n%s", string(body))

		jsonErr := json.Unmarshal(body, &apiResp)
		if jsonErr != nil {
			return errors.New("Unsuccessful on json unmarshal")
		}

		//var dishList dishDomain.Dishes
		log.Printf("Here is the apiResp.Message:%s", apiResp.Message)
		var dish dishDomain.Dish
		disherr := json.Unmarshal(apiResp.Message, &dish)
		if disherr != nil {
			return errors.New("Unsuccessful on dishlist")
		}

		speechText = "Did it work? Dish:" + dish.Title

		response.SetSimpleCard(cardTitle, speechText)
		response.SetOutputText(speechText)

	default:
		log.Println("some other intent triggered")
		speechText := "you are running some random intent"

		response.SetSimpleCard(cardTitle, speechText)
		response.SetOutputText(speechText)
	}

	return nil
}

//OnSessionEnded is a func.
func (p *ProcessEvents) OnSessionEnded(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	log.Printf("OnSessionEnded requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}

//Handle is the func run by main() which starts the program.
func Handle(context context.Context, requestEnv *alexa.RequestEnvelope) (interface{}, error) {
	return a.ProcessRequest(context, requestEnv)
}

func main() {
	lambda.Start(Handle)
}
