export class LogMessage  {
	masid: number;         	// ID of MAS agent runs in
	agentid: number;       	// ID of agent
	timestamp: Date;     	// time of message
	topic: string;         	// log type (error, debug, msg, status, app)
	msg: string;            // log message 
	data: string;				// additional information e.g in json
}