import {grpc} from '@improbable-eng/grpc-web'
import Miners from './miners'
import Asks from './asks'

export default class Client {

  private static instance: Client
  private static serviceHost: string = "some default"
  private static options?: grpc.RpcOptions

  miners: Miners
  asks: Asks

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
  }
}