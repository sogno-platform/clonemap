export interface Service  {
  id: number 
	masid: number;         		// ID of MAS agent runs in
	agentid: number;       		// ID of agent
	nodeid: number;     		  // ID of the node
	desc: string;
  createdate: Date;
  changedate: Date;
  dist: number;
}
