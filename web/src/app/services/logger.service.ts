import { Injectable } from '@angular/core';
import { WebRequestService} from './web-request.service';
import { HttpParams } from '@angular/common/http';
import { LogMessage} from 'src/app/models/logMessage.model';


@Injectable({
  providedIn: 'root'
})
export class LoggerService {

  constructor(private webReqService: WebRequestService) { }

  getAlive() {
    return this.webReqService.get('api/pf/modules')
  }

  createLoggerWithType(masid:string, agentid: string, topic: string, payload: object) {
    return this.webReqService.post(`api/logging/${masid}/${agentid}/${topic}`, payload);
  }

  createLogger(masid:string, payload: object) {
    return this.webReqService.post(`api/logging/${masid}/list`, payload);
  }

  getNLatestLogs(masid: string, agentid: string, topic:string, num: string) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/latest/${num}`);
  }
  
  getLoggerWithinRange(masid: string, agentid: string, topic: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/time/${start}/${end}`);
  }

  updateAgentState(masid: string, agentid: string, payload: object) {
    return this.webReqService.patch(`api/state/${masid}/${agentid}`, payload);
  }

  updateAgentStates(masid: string, payload: object) {
    return this.webReqService.patch(`api/state/${masid}/list`, payload);
  }
}
