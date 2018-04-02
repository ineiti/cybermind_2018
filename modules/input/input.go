package input

import "github.com/ineiti/cybermind/broker"

const ModuleInput = "input"

var InputActionPoll = broker.SubDomain("poll", ModuleInput)
