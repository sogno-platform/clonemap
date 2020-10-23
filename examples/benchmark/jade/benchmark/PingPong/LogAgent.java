package benchmark.PingPong;

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

        int numPairs;
        int numMeas;
        ACLMessage msg;
        long avg, min, max, perc;
        int[] rtts;

		public LogBehaviour(Agent a) {
            super(a);
            Object[] args = getArguments();
            numPairs = Integer.parseInt(args[0].toString());
            numMeas = 0;
            avg = 0;
            min = 1000000;
            max = 0;
            perc = 0;
            rtts = new int[numPairs*1000];
		}

		public void action() {
            msg = myAgent.receive();
            if(msg != null){
                String[] temp = msg.getContent().split(";");
                if (temp.length == 1000) {
                    for (int i = 0; i < 1000; i++) {
                        rtts[numMeas*1000+i] = Integer.parseInt(temp[i]);
                    }
                }
                numMeas++;
                if (numMeas == numPairs) {
                    Arrays.sort(rtts);
                    min = rtts[0];
                    max = rtts[numPairs*1000-1];
                    int percindex = numPairs*950+1;
                    if (percindex < numPairs*1000-1) {
                        perc = rtts[percindex];
                    }
                    for (int i = 0; i < numPairs*1000; i++) {
                        avg += rtts[i];
                    }
                    avg = avg / (numPairs*1000);
                    myLogger.log(Logger.INFO, String.valueOf(min)+","+String.valueOf(max)+","+String.valueOf(avg)+","+String.valueOf(perc));
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
