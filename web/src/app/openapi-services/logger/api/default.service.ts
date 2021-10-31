/**
 * Logger
 * API of the logger for state saving and logging
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */
/* tslint:disable:no-unused-variable member-ordering */

import { Inject, Injectable, Optional }                      from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams,
         HttpResponse, HttpEvent, HttpParameterCodec, HttpContext 
        }       from '@angular/common/http';
import { CustomHttpParameterCodec }                          from '../encoder';
import { Observable }                                        from 'rxjs';

import { LogMessage } from '../model/models';
import { LogSeries } from '../model/models';

import { BASE_PATH, COLLECTION_FORMATS }                     from '../variables';
import { Configuration }                                     from '../configuration';



@Injectable({
  providedIn: 'root'
})
export class DefaultLoggerService {

    protected basePath = 'http://localhost:30013';
    public defaultHeaders = new HttpHeaders();
    public configuration = new Configuration();
    public encoder: HttpParameterCodec;

    constructor(protected httpClient: HttpClient, @Optional()@Inject(BASE_PATH) basePath: string, @Optional() configuration: Configuration) {
        if (configuration) {
            this.configuration = configuration;
        }
        if (typeof this.configuration.basePath !== 'string') {
            if (typeof basePath !== 'string') {
                basePath = this.basePath;
            }
            this.configuration.basePath = basePath;
        }
        this.encoder = this.configuration.encoder || new CustomHttpParameterCodec();
    }


    private addToHttpParams(httpParams: HttpParams, value: any, key?: string): HttpParams {
        if (typeof value === "object" && value instanceof Date === false) {
            httpParams = this.addToHttpParamsRecursive(httpParams, value);
        } else {
            httpParams = this.addToHttpParamsRecursive(httpParams, value, key);
        }
        return httpParams;
    }

    private addToHttpParamsRecursive(httpParams: HttpParams, value?: any, key?: string): HttpParams {
        if (value == null) {
            return httpParams;
        }

        if (typeof value === "object") {
            if (Array.isArray(value)) {
                (value as any[]).forEach( elem => httpParams = this.addToHttpParamsRecursive(httpParams, elem, key));
            } else if (value instanceof Date) {
                if (key != null) {
                    httpParams = httpParams.append(key,
                        (value as Date).toISOString().substr(0, 10));
                } else {
                   throw Error("key may not be null if value is Date");
                }
            } else {
                Object.keys(value).forEach( k => httpParams = this.addToHttpParamsRecursive(
                    httpParams, value[k], key != null ? `${key}.${k}` : k));
            }
        } else if (key != null) {
            httpParams = httpParams.append(key, value);
        } else {
            throw Error("key may not be null if value is not object or array");
        }
        return httpParams;
    }

    /**
     * indicates if ams is alive
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public apiAliveGet(observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'text/plain', context?: HttpContext}): Observable<string>;
    public apiAliveGet(observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'text/plain', context?: HttpContext}): Observable<HttpResponse<string>>;
    public apiAliveGet(observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'text/plain', context?: HttpContext}): Observable<HttpEvent<string>>;
    public apiAliveGet(observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'text/plain', context?: HttpContext}): Observable<any> {

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'text/plain'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<string>(`${this.configuration.basePath}/api/alive`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

    /**
     * get logging messages within specified time range
     * @param masid ID of MAS
     * @param agentid ID of agent
     * @param name name of the log series
     * @param start start time of query
     * @param end end time of query
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public getLogSeriesInRange(masid: number, agentid: number, name: string, start: string, end: string, observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<Array<LogSeries>>;
    public getLogSeriesInRange(masid: number, agentid: number, name: string, start: string, end: string, observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpResponse<Array<LogSeries>>>;
    public getLogSeriesInRange(masid: number, agentid: number, name: string, start: string, end: string, observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpEvent<Array<LogSeries>>>;
    public getLogSeriesInRange(masid: number, agentid: number, name: string, start: string, end: string, observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<any> {
        if (masid === null || masid === undefined) {
            throw new Error('Required parameter masid was null or undefined when calling getLogSeriesInRange.');
        }
        if (agentid === null || agentid === undefined) {
            throw new Error('Required parameter agentid was null or undefined when calling getLogSeriesInRange.');
        }
        if (name === null || name === undefined) {
            throw new Error('Required parameter name was null or undefined when calling getLogSeriesInRange.');
        }
        if (start === null || start === undefined) {
            throw new Error('Required parameter start was null or undefined when calling getLogSeriesInRange.');
        }
        if (end === null || end === undefined) {
            throw new Error('Required parameter end was null or undefined when calling getLogSeriesInRange.');
        }

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'application/json'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<Array<LogSeries>>(`${this.configuration.basePath}/api/logging/series/${encodeURIComponent(String(masid))}/${encodeURIComponent(String(agentid))}/${encodeURIComponent(String(name))}/time/${encodeURIComponent(String(start))}/${encodeURIComponent(String(end))}`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

    /**
     * get the names of the logSeries
     * @param masid ID of MAS
     * @param agentid ID of agent
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public getLogSeriesNames(masid: number, agentid: number, observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<Array<string>>;
    public getLogSeriesNames(masid: number, agentid: number, observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpResponse<Array<string>>>;
    public getLogSeriesNames(masid: number, agentid: number, observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpEvent<Array<string>>>;
    public getLogSeriesNames(masid: number, agentid: number, observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<any> {
        if (masid === null || masid === undefined) {
            throw new Error('Required parameter masid was null or undefined when calling getLogSeriesNames.');
        }
        if (agentid === null || agentid === undefined) {
            throw new Error('Required parameter agentid was null or undefined when calling getLogSeriesNames.');
        }

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'application/json'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<Array<string>>(`${this.configuration.basePath}/api/logging/series/${encodeURIComponent(String(masid))}/${encodeURIComponent(String(agentid))}/names`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

    /**
     * get logging messages within specified time range
     * @param masid ID of MAS
     * @param agentid ID of agent
     * @param topic type of logging
     * @param start start time of query
     * @param end end time of query
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public getLogsInRange(masid: number, agentid: number, topic: 'error' | 'debug' | 'msg' | 'status' | 'app', start: string, end: string, observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<Array<LogMessage>>;
    public getLogsInRange(masid: number, agentid: number, topic: 'error' | 'debug' | 'msg' | 'status' | 'app', start: string, end: string, observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpResponse<Array<LogMessage>>>;
    public getLogsInRange(masid: number, agentid: number, topic: 'error' | 'debug' | 'msg' | 'status' | 'app', start: string, end: string, observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpEvent<Array<LogMessage>>>;
    public getLogsInRange(masid: number, agentid: number, topic: 'error' | 'debug' | 'msg' | 'status' | 'app', start: string, end: string, observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<any> {
        if (masid === null || masid === undefined) {
            throw new Error('Required parameter masid was null or undefined when calling getLogsInRange.');
        }
        if (agentid === null || agentid === undefined) {
            throw new Error('Required parameter agentid was null or undefined when calling getLogsInRange.');
        }
        if (topic === null || topic === undefined) {
            throw new Error('Required parameter topic was null or undefined when calling getLogsInRange.');
        }
        if (start === null || start === undefined) {
            throw new Error('Required parameter start was null or undefined when calling getLogsInRange.');
        }
        if (end === null || end === undefined) {
            throw new Error('Required parameter end was null or undefined when calling getLogsInRange.');
        }

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'application/json'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<Array<LogMessage>>(`${this.configuration.basePath}/api/logging/${encodeURIComponent(String(masid))}/${encodeURIComponent(String(agentid))}/${encodeURIComponent(String(topic))}/time/${encodeURIComponent(String(start))}/${encodeURIComponent(String(end))}`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

    /**
     * get heatmap of the messages
     * @param masid ID of MAS
     * @param start start time of query
     * @param end end time of query
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public getMsgHeatmap(masid: number, start: string, end: string, observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<Array<string>>;
    public getMsgHeatmap(masid: number, start: string, end: string, observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpResponse<Array<string>>>;
    public getMsgHeatmap(masid: number, start: string, end: string, observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpEvent<Array<string>>>;
    public getMsgHeatmap(masid: number, start: string, end: string, observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<any> {
        if (masid === null || masid === undefined) {
            throw new Error('Required parameter masid was null or undefined when calling getMsgHeatmap.');
        }
        if (start === null || start === undefined) {
            throw new Error('Required parameter start was null or undefined when calling getMsgHeatmap.');
        }
        if (end === null || end === undefined) {
            throw new Error('Required parameter end was null or undefined when calling getMsgHeatmap.');
        }

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'application/json'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<Array<string>>(`${this.configuration.basePath}/logging/stats/${encodeURIComponent(String(masid))}/heatmap/${encodeURIComponent(String(start))}/${encodeURIComponent(String(end))}`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

    /**
     * get stats of the behaviors
     * @param masid ID of MAS
     * @param agentid ID of agent
     * @param behtype type of the behavior
     * @param start start time of query
     * @param end end time of query
     * @param observe set whether or not to return the data Observable as the body, response or events. defaults to returning the body.
     * @param reportProgress flag to report request and response progress.
     */
    public getStats(masid: number, agentid: number, behtype: string, start: string, end: string, observe?: 'body', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<Array<LogMessage>>;
    public getStats(masid: number, agentid: number, behtype: string, start: string, end: string, observe?: 'response', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpResponse<Array<LogMessage>>>;
    public getStats(masid: number, agentid: number, behtype: string, start: string, end: string, observe?: 'events', reportProgress?: boolean, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<HttpEvent<Array<LogMessage>>>;
    public getStats(masid: number, agentid: number, behtype: string, start: string, end: string, observe: any = 'body', reportProgress: boolean = false, options?: {httpHeaderAccept?: 'application/json', context?: HttpContext}): Observable<any> {
        if (masid === null || masid === undefined) {
            throw new Error('Required parameter masid was null or undefined when calling getStats.');
        }
        if (agentid === null || agentid === undefined) {
            throw new Error('Required parameter agentid was null or undefined when calling getStats.');
        }
        if (behtype === null || behtype === undefined) {
            throw new Error('Required parameter behtype was null or undefined when calling getStats.');
        }
        if (start === null || start === undefined) {
            throw new Error('Required parameter start was null or undefined when calling getStats.');
        }
        if (end === null || end === undefined) {
            throw new Error('Required parameter end was null or undefined when calling getStats.');
        }

        let localVarHeaders = this.defaultHeaders;

        let localVarHttpHeaderAcceptSelected: string | undefined = options && options.httpHeaderAccept;
        if (localVarHttpHeaderAcceptSelected === undefined) {
            // to determine the Accept header
            const httpHeaderAccepts: string[] = [
                'application/json'
            ];
            localVarHttpHeaderAcceptSelected = this.configuration.selectHeaderAccept(httpHeaderAccepts);
        }
        if (localVarHttpHeaderAcceptSelected !== undefined) {
            localVarHeaders = localVarHeaders.set('Accept', localVarHttpHeaderAcceptSelected);
        }

        let localVarHttpContext: HttpContext | undefined = options && options.context;
        if (localVarHttpContext === undefined) {
            localVarHttpContext = new HttpContext();
        }


        let responseType_: 'text' | 'json' = 'json';
        if(localVarHttpHeaderAcceptSelected && localVarHttpHeaderAcceptSelected.startsWith('text')) {
            responseType_ = 'text';
        }

        return this.httpClient.get<Array<LogMessage>>(`${this.configuration.basePath}/api/logging/stats/${encodeURIComponent(String(masid))}/${encodeURIComponent(String(agentid))}/${encodeURIComponent(String(behtype))}/${encodeURIComponent(String(start))}/${encodeURIComponent(String(end))}`,
            {
                context: localVarHttpContext,
                responseType: <any>responseType_,
                withCredentials: this.configuration.withCredentials,
                headers: localVarHeaders,
                observe: observe,
                reportProgress: reportProgress
            }
        );
    }

}
