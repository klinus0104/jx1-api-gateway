if db_id('gm_api_test') is null create database gm_api_test;
go
use gm_api_test;
if object_id('Account_info','U') is null
begin
  create table Account_info (cAccName nvarchar(64) not null primary key, iClientID int not null default 0, nUserIP bigint not null default 0, cPassword nvarchar(128) null, cSecPassword nvarchar(128) null);
end;
if not exists(select 1 from Account_info where cAccName='gm_fixture_account')
  insert Account_info(cAccName,iClientID,nUserIP,cPassword) values('gm_fixture_account',42,2130706433,'fixture-password');
