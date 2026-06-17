import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';

@Injectable()
export class InternalApiKeyGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest();
    const apiKey = request.headers['x-internal-api-key'];

    const expectedApiKey = process.env.INTERNAL_API_KEY ?? 'local_internal_key';

    if (!apiKey || apiKey !== expectedApiKey) {
      throw new UnauthorizedException('Invalid internal API key');
    }

    return true;
  }
}
