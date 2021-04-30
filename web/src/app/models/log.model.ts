export class LogMessage  {
	masid: number;         	// ID of MAS agent runs in
	agentid: number;       	// ID of agent
	timestamp: Date;     	// time of message
	topic: string;         	// log type (error, debug, msg, status, app)
	msg: string;            // log message 
	data: string;				// additional information e.g in json
}

export class LogSeries  {
	masid: number;         		// ID of MAS agent runs in
	agentid: number;       		// ID of agent
	timestamp: Date;     		// time of the log series
	name: string;            	// name of the log series
	value: number;				// value of the log series
}


export class pointSeries {
	name: string;
	series: point[];
}

class point {
	name: string;
	x: number;
	y: number;
	r: number;
}



