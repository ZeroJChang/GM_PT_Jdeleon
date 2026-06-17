import {
  Body,
  Controller,
  Get,
  Param,
  Post,
  UseGuards,
} from '@nestjs/common';

import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { Roles } from '../auth/decorators/roles.decorator';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import type { AuthenticatedUser } from '../auth/types/authenticated-user.type';
import { CreateLoanDto } from './dto/create-loan.dto';
import { LoansService } from './loans.service';

@Controller('loans')
@UseGuards(JwtAuthGuard, RolesGuard)
export class LoansController {
  constructor(private readonly loansService: LoansService) {}

  @Post()
  createLoan(
    @CurrentUser() currentUser: AuthenticatedUser,
    @Body() createLoanDto: CreateLoanDto,
  ) {
    return this.loansService.createLoan(currentUser, createLoanDto);
  }

  @Post(':id/return')
  returnLoan(@Param('id') id: string) {
    return this.loansService.returnLoan(id);
  }

  @Get('me/active')
  findMyActiveLoans(@CurrentUser() currentUser: AuthenticatedUser) {
    return this.loansService.findMyActiveLoans(currentUser);
  }

  @Get('history')
  @Roles('ADMIN')
  findHistory(@CurrentUser() currentUser: AuthenticatedUser) {
    return this.loansService.findHistory(currentUser);
  }
}
