package main

import (
  "fmt"
  "github.com/Jeffail/gabs"
  "github.com/parnurzeal/gorequest"
  "hash/fnv"
  "math/rand"
  "os"
  "strings"
  "time"
)

func currentEventId(apiKey string, meetupName string) string {
  currentTime := time.Now().UTC().Unix()
  
  request := gorequest.New()
  _, body, _ := request.Get("https://api.meetup.com/" + meetupName + "/events?&key=" + apiKey + "&page=5").End()

  jsonParsed, _ := gabs.ParseJSON([]byte(body))
  children, _ := jsonParsed.Children()
  eventTime, _ := children[0].Path("time").Data().(float64)
  timeToEvent := ((int64(eventTime) / int64(1000)) - currentTime)
  if int64(0) <= timeToEvent && timeToEvent <= int64(1800) {
    eventId, _ := children[0].Path("id").Data().(string)
    
    return eventId
  } else {
    
    return "false"
  }
}

func getNames(apiKey string, eventId string) []string {
  request := gorequest.New()
  _, body, _ := request.Get("https://api.meetup.com/2/rsvps?key=" + apiKey + "&event_id=" + eventId + "&page=2000").End()

  jsonParsed, _ := gabs.ParseJSON([]byte(body))
  children, _ := jsonParsed.Path("results").Children()
  
  var names []string
  for _, child := range children {
    if child.Path("response").Data().(string) == "yes" {
        names = append(names, child.Path("member.name").Data().(string))
    }
  }
  
  return names
}

func grouper(length int) int {
	var groups = 2
  var members = length/groups
  
  for members >= 10 {
    groups = groups + 2
    members = length/groups
  }
  
  return groups
}

func teamSpliter(names []string, randOrder []int, length int, groups int) [][]string {
  teams := make([][]string, groups)
  
  var j = 0
  
  for _, index := range randOrder {
    if j >= groups {
      j = 0
    }
    teams[j] = append(teams[j], names[index])
    j += 1
  }
  
  return teams
}

func teamStringify(teams [][]string) []string {
  var teamStrings []string
  i := 1
  for _, team := range teams {
    var teamString = fmt.Sprintf("%s%d:\n", "Team ", i)
    for _, member := range team {
      teamString += member + "\n"
    }
    teamStrings = append(teamStrings, teamString)
    i += 1
  }
  
  return teamStrings
}

func getEventCommentId(apiKey string, meetupName string, eventId string) string {
  request := gorequest.New()
  _, body, _ := request.Get("https://api.meetup.com/" + meetupName + "/events/" + eventId + "/comments?key=" + apiKey).End()
  
  jsonParsed, _ := gabs.ParseJSON([]byte(body))
  children, _ := jsonParsed.Children()
  
  var commentId = "0"
  for _, child := range children {
    if child.Path("member.name").Data().(string) == "Robert Beatty" {
      commentId = fmt.Sprint(int(child.Path("id").Data().(float64)))
    }
  }
  
  return commentId
}

func deleteEventComment(apiKey string, meetupName string, eventId string, commentId string) {
  if commentId != "0" {
    request := gorequest.New()
    request.Delete("https://api.meetup.com/" + meetupName + "/events/" + eventId + "/comments/comment:" + commentId + "?key=" + apiKey).End()
  }
}

func sendEventComment(apiKey string, eventId string, teamStrings []string) {
  comment := fmt.Sprintf(strings.Join(teamStrings, "\n"))
  request := gorequest.New()
  request.Post("https://api.meetup.com/2/event_comment?key=" + apiKey + "&event_id=" + eventId + "&comment=" + comment).End()
}

func hashEventId(eventId string) int64 {
      hash := fnv.New32a()
      hash.Write([]byte(eventId))
      return int64(hash.Sum32())
}

func main() {
  apiKey := os.Getenv("MEETUPAPIKEY")
  meetupName := "itpfootballclub"

  eventId := currentEventId(apiKey, meetupName)
  fmt.Println(eventId)
  if eventId == "false" {
  } else {
    names := getNames(apiKey, eventId)
    length := len(names)
    rand.Seed(hashEventId(eventId))
    randOrder := rand.Perm(length)
    groups := grouper(length)
    teams := teamSpliter(names, randOrder, length, groups)
    teamStrings := teamStringify(teams)
    fmt.Println(teamStrings)
    commentId := getEventCommentId(apiKey, meetupName, eventId)
    deleteEventComment(apiKey, meetupName, eventId, commentId)
    sendEventComment(apiKey, eventId, teamStrings)
  }
}
