import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http'
import { environment } from 'src/environments/environment.prod'
@Injectable({
  providedIn: 'root'
})
export class WebRequestService {
    readonly ROOT_URL = environment.gateway;

    constructor(private http: HttpClient) {
    }

    get(uri: string) {
        return this.http.get(`${this.ROOT_URL}/${uri}`);
    }

    getText(uri: string) {
        return this.http.get(`${this.ROOT_URL}/${uri}`, {responseType: 'text'})
    }


    post(uri: string, payload: object) {
        return this.http.post<any>(`${this.ROOT_URL}/${uri}`, payload);
    }

    patch(uri: string, payload: Object) {
        return this.http.patch(`${this.ROOT_URL}/${uri}`, payload);
    }

    delete(uri: string) {
        return this.http.delete(`${this.ROOT_URL}/${uri}`);
    }

}
