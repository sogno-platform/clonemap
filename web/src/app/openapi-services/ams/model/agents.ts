/**
 * Agent Management System
 * API of the Agent Management System for user interaction with the MAS and MAS-internal communication
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */
import { AgentInfo } from './agentInfo';


/**
 * contains informaton about how many agents are running
 */
export interface Agents { 
    /**
     * number of running agents
     */
    counter: number;
    /**
     * all agents in mas
     */
    instances: Array<AgentInfo>;
}

