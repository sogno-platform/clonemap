package benchmark.RepPingPong;

import jade.core.*;
import jade.core.behaviours.*;
import jade.lang.acl.ACLMessage;
import jade.domain.FIPAAgentManagement.ServiceDescription;
import jade.domain.FIPAAgentManagement.DFAgentDescription;
import jade.core.replication.AgentReplicationHandle;
import jade.core.replication.AgentReplicationHelper;
import jade.domain.DFService;
import jade.domain.FIPAException;
import jade.wrapper.ControllerException;
import jade.util.Logger;
import java.util.Random;

/**
 */
public class PingPong extends Agent implements AgentReplicationHelper.Listener {
// public class PingPong extends Agent {

	private Logger myLogger = Logger.getMyLogger(getClass().getName());
    private int state = 0;

	private class PingPongBehaviour extends CyclicBehaviour {

        ACLMessage msg;
        ACLMessage msg2;
        private boolean finished = false;
        long startTime;
        int counter = 0;
        long[] rtts;
        boolean start = false;
        long timeStateUpdate = 0;

		public PingPongBehaviour(Agent a) {
            super(a);
            try{
                Thread.sleep(10000);
            }
            catch (InterruptedException temp){
            }
            if (!getLocalName().endsWith("_R")) {
                 Random rand = new Random();
                // Obtain a number between [0 - 49].
                int n = rand.nextInt(12);
                n += 1;
                try {
                     n = Integer.parseInt(getContainerController().getContainerName().split("jade")[1]);
                }
                catch (ControllerException temp){
                }
                n += 1;
                if (n > 12) {
                    n = 1;
                }
                createReplica(getLocalName()+"_R", "jade"+String.valueOf(n));
            }
            Object[] args = getArguments();
            msg2 = new ACLMessage(ACLMessage.INFORM);
            msg2.setSender(myAgent.getAID());
            AID receiver = new AID(args[0].toString()+"_V", AID.ISLOCALNAME);
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
            timeStateUpdate += elapsedTime;
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
                    if (start && timeStateUpdate > 25000) {
                        setState(state+1);
                        timeStateUpdate = 0;
                    }
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
                        setState(-1);
                    }
                }
            }
            else {
                block();
            }
        }
        
	}

    @Override
	protected void setup() {
        try {
            // Makes this agent become the master replica of a newly defined replicated agent
            AgentReplicationHelper helper = (AgentReplicationHelper) getHelper(AgentReplicationHelper.SERVICE_NAME);
            AID virtualAid = helper.makeVirtual(getLocalName()+"_V", AgentReplicationHelper.COLD_REPLICATION);
        }
        catch (ServiceException se) {
			System.out.println("Agent "+getLocalName()+" - Error retrieving AgentReplicationHelper!!! Check that the AgentReplicationService is correctly installed in this container");
			se.printStackTrace();
			doDelete();
		}
		PingPongBehaviour PPBehaviour = new  PingPongBehaviour(this);
		addBehaviour(PPBehaviour);
	}

	void createReplica(String replicaName, String where) {
		if (replicaName == null || replicaName.trim().length() == 0) {
			System.out.println("Replica name not specified");
			return;
		}
		if (where == null || where.trim().length() == 0) {
			System.out.println("Replica location not specified");
			return;
		}
		try {
			AgentReplicationHelper helper = (AgentReplicationHelper) getHelper(AgentReplicationHelper.SERVICE_NAME);
			helper.createReplica(replicaName.trim(), new ContainerID(where.trim(), null));
		}
		catch (Exception e) {
			System.out.println("Agent "+getLocalName()+" - Error creating replica on container "+where);
			e.printStackTrace();
		}
	}

	public void setState(int newValue) {
		// The call to setValue() will be invoked on other replicas too 
		AgentReplicationHandle.replicate(this, "setState", new Object[]{newValue});
		state = newValue;
	}

    @Override
	public void replicaAdded(AID replicaAid, Location where) {
	}

	@Override
	public void replicaRemoved(AID replicaAid, Location where) {
	}

	@Override
	public void replicaCreationFailed(AID replicaAid, Location where) {
		System.out.println("Agent "+getLocalName()+" - Creation of new replica "+replicaAid.getLocalName()+" in "+where.getName()+" failed");
	}

	@Override
	public void becomeMaster() {
	}
}
