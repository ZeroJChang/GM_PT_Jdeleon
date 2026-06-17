import {
  BadGatewayException,
  ForbiddenException,
  Injectable,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

import { AuthenticatedUser } from '../auth/types/authenticated-user.type';
import { CreateLoanDto } from './dto/create-loan.dto';

@Injectable()
export class LoansService {
  private readonly loansServiceUrl: string;

  constructor(private readonly configService: ConfigService) {
    this.loansServiceUrl =
      this.configService.get<string>('LOANS_SERVICE_URL') ??
      'http://localhost:8080';
  }

  async createLoan(currentUser: AuthenticatedUser, createLoanDto: CreateLoanDto) {
    const userId =
      currentUser.role === 'ADMIN' && createLoanDto.userId
        ? createLoanDto.userId
        : currentUser.sub;

    return this.requestToLoansService('/loans', {
      method: 'POST',
      body: {
        userId,
        bookId: createLoanDto.bookId,
      },
    });
  }

  async returnLoan(loanId: string) {
    return this.requestToLoansService(`/loans/${loanId}/return`, {
      method: 'POST',
    });
  }

  async findMyActiveLoans(currentUser: AuthenticatedUser) {
    return this.requestToLoansService(`/loans/users/${currentUser.sub}/active`, {
      method: 'GET',
    });
  }

  async findHistory(currentUser: AuthenticatedUser) {
    if (currentUser.role !== 'ADMIN') {
      throw new ForbiddenException('Only admin users can access loan history');
    }

    return this.requestToLoansService('/loans/history', {
      method: 'GET',
    });
  }

  private async requestToLoansService(
    path: string,
    options: {
      method: 'GET' | 'POST';
      body?: unknown;
    },
  ) {
    const response = await fetch(`${this.loansServiceUrl}${path}`, {
      method: options.method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
    });

    const responseBody = await response.text();

    let data: unknown = null;

    if (responseBody) {
      try {
        data = JSON.parse(responseBody);
      } catch {
        data = {
          message: responseBody,
        };
      }
    }

    if (!response.ok) {
      throw new BadGatewayException({
        message: 'Loans service request failed',
        statusCode: response.status,
        error: data,
      });
    }

    return data;
  }
}
