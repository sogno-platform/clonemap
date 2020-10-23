package benchmark.CPULoad;

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
public class CPULoad extends Agent {

	private Logger myLogger = Logger.getMyLogger(getClass().getName());

	private class CPULoadBehaviour extends CyclicBehaviour {
        ACLMessage msg;
        ACLMessage msg2;
        long startTime;
        long lastSentTime;
        double[] a;
        double[] b;
        double[] c;
        float T;
        float Tr;

		public CPULoadBehaviour(Agent ag) {
            super(ag);
            Object[] args = getArguments();
            T = Float.parseFloat(args[2].toString());
            Tr = Float.parseFloat(args[3].toString());
            a = new double[100];
            b = new double[100];
            c = new double[100];
            for (int i = 0; i < 100; i++) {
                a[i] = Math.random();
                b[i] = Math.random();
                c[i] = Math.random();
            }
            msg2 = new ACLMessage(ACLMessage.INFORM);
            msg2.setSender(myAgent.getAID());
            AID receiver = new AID(args[0].toString(), AID.ISLOCALNAME);
            msg2.addReceiver(receiver);
            msg2.setProtocol("p");
            msg2.setContent("hello world");
            lastSentTime = 0;
            try{
                Thread.sleep(80000);
            }
            catch (InterruptedException temp){
            }
		}

		public void action() {
            startTime = System.nanoTime();
            for (;;) {
                for (int i = 0; i < 100; i++) {
                    a[i] = b[i] + c[i];
                }
                float diff = (System.nanoTime() - startTime) / 1000000;
                if (diff > T*Tr) {
                    break;
                }
            }
            if (System.nanoTime() > lastSentTime+1000000) {
                lastSentTime = System.nanoTime();
                send(msg2);
                msg = myAgent.receive();
            }
            try{
                Thread.sleep((long)(T*(1-Tr)));
            }
            catch (InterruptedException temp){
            }
        }
        
    }

	protected void setup() {
		CPULoadBehaviour CPUBehaviour = new  CPULoadBehaviour(this);
        addBehaviour(CPUBehaviour);
	}
}
