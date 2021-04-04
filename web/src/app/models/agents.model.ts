export class Agents {
	counter: number;    
	instances:    AgentInfo[];
}

class AgentInfo {
	spec:         AgentSpec;
	masid:        number;        // ID of MAS
	agencyid:     number;        // name of the agency
	imid: 		  number;           // ID of agency image
	id:           number;        // ID of agent
	address:      Address;  
	status:       Status;     
}

class AgentSpec {
	NodeID   :number;              // id of the node the agent is attached to
	Name     :string;  // name/description of agent
	AType    :string;  // type of agent (application dependent)
	ASubtype :string;  // subtype of agent (application dependent)
	Custom   :string;  // custom configuration data
}
class Address {
	Agency :string;
}

class Status {
	Code:       number;     // status code
	LastUpdate: Date;       // time of last update
}