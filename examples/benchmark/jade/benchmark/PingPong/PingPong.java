package benchmark.PingPong;

import jade.core.*;
import jade.core.behaviours.*;
import jade.lang.acl.ACLMessage;
import jade.domain.FIPAAgentManagement.ServiceDescription;
import jade.domain.FIPAAgentManagement.DFAgentDescription;
import jade.domain.DFService;
import jade.domain.FIPAException;
import jade.util.Logger;

/**
 */
public class PingPong extends Agent {

	private Logger myLogger = Logger.getMyLogger(getClass().getName());

	private class PingPongBehaviour extends CyclicBehaviour {

        ACLMessage msg;
        ACLMessage msg2;
        private boolean finished = false;
        long startTime;
        int counter = 0;
        long[] rtts;
        boolean start = false;

		public PingPongBehaviour(Agent a) {
            super(a);
            Object[] args = getArguments();
            msg2 = new ACLMessage(ACLMessage.INFORM);
            msg2.setSender(myAgent.getAID());
            AID receiver = new AID(args[0].toString(), AID.ISLOCALNAME);
            msg2.addReceiver(receiver);
            msg2.setProtocol("p");
            msg2.setContent("hello world");
            rtts = new long[1000];
            try{
                Thread.sleep(10000);
            }
            catch (InterruptedException temp){
            }
            startTime = System.nanoTime();
            if(args[1].toString().equals("true")) {
                start = true;
                send(msg2);
            }
		}

		public void action() {
            msg = myAgent.receive();
            long stopTime = System.nanoTime();
            long elapsedTime = (stopTime - startTime)/1000;
            if(msg != null){
                msg = null;
                if (counter < 10000) {
                    if (counter >= 1000 && counter < 2000) {
                        rtts[counter-1000] = elapsedTime;
                    }
                    msg2.setContent(String.valueOf(counter));
                    startTime = System.nanoTime();
                    send(msg2);
                    counter++;
                }
                if (counter == 3000) {
                    String content = "";
                    long avg, min, max;
                    min = 1000;
                    max = 0;
                    avg = 0;
                    for (int i = 0; i < 1000; i++) {
                        content += String.valueOf(rtts[i]);
                        if (i < 999) {
                            content+= ";";
                        }
                        if (rtts[i] > max) {
                            max = rtts[i];
                        }
                        if (rtts[i] < min) {
                            min = rtts[i];
                        }
                        avg += rtts[i];
                    }
                    avg = avg / 1000;
                    ACLMessage logmsg = new ACLMessage(ACLMessage.INFORM);
                    logmsg.setSender(myAgent.getAID());
                    AID receiver = new AID("logag", AID.ISLOCALNAME);
                    logmsg.addReceiver(receiver);
                    logmsg.setProtocol("p");

                    logmsg.setContent(content);
                    if(start) {
                        send(logmsg);
                    }
                }
            }
            else {
                block();
            }
        }
        
	}


	protected void setup() {
		PingPongBehaviour PPBehaviour = new  PingPongBehaviour(this);
		addBehaviour(PPBehaviour);
	}
}
