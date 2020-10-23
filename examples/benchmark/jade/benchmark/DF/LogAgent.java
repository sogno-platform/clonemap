package benchmark.DF;

import jade.core.*;
import jade.core.behaviours.*;
import jade.lang.acl.ACLMessage;
import jade.domain.FIPAAgentManagement.ServiceDescription;
import jade.domain.FIPAAgentManagement.DFAgentDescription;
import jade.domain.DFService;
import jade.domain.FIPAException;
import jade.util.Logger;
import java.util.Arrays;

/**
 */
public class LogAgent extends Agent {

	private Logger myLogger = Logger.getMyLogger(getClass().getName());

	private class LogBehaviour extends CyclicBehaviour {

        int numAgs;
        int numMeas;
        ACLMessage msg;
        long avgSum, minSum, maxSum, percSum, avgReg, minReg, maxReg, percReg, avgSearch, minSearch, maxSearch, percSearch, avgDereg, minDereg, maxDereg, percDereg;
        int[] rttsSum, rttsReg, rttsSearch, rttsDereg;

		public LogBehaviour(Agent a) {
            super(a);
            Object[] args = getArguments();
            numAgs = Integer.parseInt(args[0].toString());
            avgSum = 0;
            minSum = 1000000;
            maxSum = 0;
            percSum = 0;
            avgReg = 0;
            minReg = 1000000;
            maxReg = 0;
            percReg = 0;
            avgSearch = 0;
            minSearch = 1000000;
            maxSearch = 0;
            percSearch = 0;
            avgDereg = 0;
            minDereg = 1000000;
            maxDereg = 0;
            percDereg = 0;
            rttsSum = new int[numAgs];
            rttsReg = new int[numAgs];
            rttsSearch = new int[numAgs*8];
            rttsDereg = new int[numAgs];
            numMeas = 0;
		}

		public void action() {
            msg = myAgent.receive();
            if(msg != null){
                String[] temp = msg.getContent().split(";");
                if (temp.length == 11) {
                    for (int i = 0; i < 11; i++) {
                        if (i == 0) {
                            rttsReg[numMeas] = Integer.parseInt(temp[i]);
                        } else if (i == 9) {
                            rttsDereg[numMeas] = Integer.parseInt(temp[i]);
                        } else if (i == 10) {
                            rttsSum[numMeas] = Integer.parseInt(temp[i]);
                        } else {
                            rttsSearch[numMeas*8+(i-1)] = Integer.parseInt(temp[i]);
                        }
                    }
                }
                numMeas++;
                if (numMeas == numAgs) {
                    Arrays.sort(rttsSum);
                    minSum = rttsSum[0];
                    maxSum = rttsSum[numAgs-1];
                    int percindex = numAgs*95/100;
                    if (percindex < numAgs) {
                        percSum = rttsSum[percindex];
                    }
                    for (int i = 0; i < numAgs; i++) {
                        avgSum += rttsSum[i];
                    }
                    avgSum = avgSum / numAgs;
                    myLogger.log(Logger.INFO, "Sum: "+ String.valueOf(minSum)+","+String.valueOf(maxSum)+","+String.valueOf(avgSum)+","+String.valueOf(percSum));

                    Arrays.sort(rttsReg);
                    minReg = rttsReg[0];
                    maxReg = rttsReg[numAgs-1];
                    percindex = numAgs*95/100;
                    if (percindex < numAgs) {
                        percReg = rttsReg[percindex];
                    }
                    for (int i = 0; i < numAgs; i++) {
                        avgReg += rttsReg[i];
                    }
                    avgReg = avgReg / numAgs;
                    myLogger.log(Logger.INFO, "Reg: "+ String.valueOf(minReg)+","+String.valueOf(maxReg)+","+String.valueOf(avgReg)+","+String.valueOf(percReg));

                    Arrays.sort(rttsSearch);
                    minSearch = rttsSearch[0];
                    maxSearch = rttsSearch[8*numAgs-1];
                    percindex = numAgs*8*95/100;
                    if (percindex < numAgs*8) {
                        percSearch = rttsSearch[percindex];
                    }
                    for (int i = 0; i < 8*numAgs; i++) {
                        avgSearch += rttsSearch[i];
                    }
                    avgSearch = avgSearch / (8*numAgs);
                    myLogger.log(Logger.INFO, "Search: "+ String.valueOf(minSearch)+","+String.valueOf(maxSearch)+","+String.valueOf(avgSearch)+","+String.valueOf(percSearch));

                    Arrays.sort(rttsDereg);
                    minDereg = rttsDereg[0];
                    maxDereg = rttsDereg[numAgs-1];
                    percindex = numAgs*95/100;
                    if (percindex < numAgs) {
                        percDereg = rttsDereg[percindex];
                    }
                    for (int i = 0; i < numAgs; i++) {
                        avgDereg += rttsDereg[i];
                    }
                    avgDereg = avgDereg / numAgs;
                    myLogger.log(Logger.INFO, "Dereg: "+ String.valueOf(minDereg)+","+String.valueOf(maxDereg)+","+String.valueOf(avgDereg)+","+String.valueOf(percDereg));
                }
            }
            else {
                block();
            }
        }
        
	}


	protected void setup() {
		LogBehaviour LBehaviour = new  LogBehaviour(this);
		addBehaviour(LBehaviour);
	}
}
