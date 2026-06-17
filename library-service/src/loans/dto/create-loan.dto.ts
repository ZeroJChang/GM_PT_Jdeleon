import { IsNotEmpty, IsOptional, IsString } from 'class-validator';

export class CreateLoanDto {
  @IsString()
  @IsNotEmpty()
  bookId: string;

  @IsOptional()
  @IsString()
  userId?: string;
}
