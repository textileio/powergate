import {grpc} from '@improbable-eng/grpc-web'
import Miners from './miners'
import Asks from './asks'
import Slashing from './slashing'
import Reputation from './reputation'

export default class Client {

  private static instance: Client
  private static serviceHost: string = "some default"
  private static options?: grpc.RpcOptions

  miners: Miners
  asks: Asks
  slashing: Slashing
  reputation: Reputation

  static initialize(serviceHost: string, options?: grpc.RpcOptions) {
    Client.serviceHost = serviceHost
    Client.options = options
  }

  static shared(): Client {
    if (!Client.instance) {
        Client.instance = new Client()
    }
    return Client.instance
  }

  private constructor() {
    this.miners = new Miners(Client.serviceHost, Client.options)
    this.asks = new Asks(Client.serviceHost, Client.options)
    this.slashing = new Slashing(Client.serviceHost, Client.options)
    this.reputation = new Reputation(Client.serviceHost, Client.options)
  }
}