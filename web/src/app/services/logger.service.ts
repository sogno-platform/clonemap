import { Injectable } from '@angular/core';
import { WebRequestService} from './web-request.service'

@Injectable({
  providedIn: 'root'
})
export class LoggerService {

  constructor(private webReqService: WebRequestService) { }

  getAlive() {
    return this.webReqService.getText('alive/logger')
  }

  createLoggerWithType(masid:number, agentid: number, topic: string, payload: object) {
    return this.webReqService.post(`api/logging/${masid}/${agentid}/${topic}`, payload);
  }

  createLogger(masid:number, payload: object) {
    return this.webReqService.post(`api/logging/${masid}/list`, payload);
  }

  getNLatestLogger(masid: number, agentid: number, topic:string, num: number) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/latest/${num}`);
  }
  
  getLoggerWithinRange(masid: number, agentid: number, topic: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/time/${start}/${end}`);
  }

  updateAgentState(masid: number, agentid: number, payload: object) {
    return this.webReqService.patch(`api/state/${masid}/${agentid}`, payload);
  }

  updateAgentStates(masid: number, payload: object) {
    return this.webReqService.patch(`api/state/${masid}/list`, payload);
  }
}
