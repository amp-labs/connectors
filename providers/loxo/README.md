# Loxo connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects


| Object                                    | Resource                                 | Method     |
| ----------------------------------------- | ---------------------------------------- | ---------- |
| activity_types                            | activity_types                           | read       |
| address_types                             | address_types                            | read       |
| bonus_payment_types                       | bonus_payment_types                      | read       |
| bonus_types                               | bonus_types                              | read       |
| companies                                 | companies                                | read,write |
| company_global_statuses                   | company_global_statuses                  | read       |
| company_types                             | company_types                            | read       |
| compensation_types                        | compensation_types                       | read       |
| countries                                 | countries                                | read       |
| currencies                                | currencies                               | read       |
| deal_workflows                            | deal_workflows                           | read       |
| deals                                     | deals                                    | read,write |
| disability_statuses                       | disability_statuses                      | read       |
| diversity_types                           | diversity_types                          | read       |
| dynamic_fields                            | dynamic_fields                           | read       |
| education_types                           | education_types                          | read       |
| email_tracking                            | email_tracking                           | read       |
| email_types                               | email_types                              | read       |
| equity_types                              | equity_types                             | read       |
| ethnicities                               | ethnicities                              | read       |
| fee_types                                 | fee_types                                | read       |
| form_templates                            | form_templates                           | read       |
| forms                                     | forms                                    | read,write |
| genders                                   | genders                                  | read       |
| job_categories                            | job_categories                           | read       |
| job_contact_types                         | job_contact_types                        | read       |
| job_owner_types                           | job_owner_types                          | read       |
| job_statuses                              | job_statuses                             | read       |
| job_types                                 | job_types                                | read       |
| jobs                                      | jobs                                     | read,write |
| people                                    | people                                   | read,write |
| person_events                             | person_events                            | read,write |
| person_global_statuses                    | person_global_statuses                   | read       |
| person_lists                              | person_lists                             | read       |
| person_share_field_types                  | person_share_field_types                 | read       |
| person_types                              | person_types                             | read       |
| phone_types                               | phone_types                              | read       |
| placements                                | placements                               | read,write |
| pronouns                                  | pronouns                                 | read       |
| question_types                            | question_types                           | read       |
| schedule_items                            | schedule_items                           | read       |
| scorecards                                | scorecards                               | read,write |
| scorecards/scorecard_recommendation_types | scorecards/scorecard_recommendation_types| read       |
| scorecards/scorecard_types                | scorecards/scorecard_types               | read       |
| scorecards/scorecard_templates            | scorecards/scorecard_templates           | read,write |
| scorecards/scorecard_visibility_types     | scorecards/scorecard_visibility_types    | read       |
| seniority_levels                          | seniority_levels                         | read       |
| sms                                       | sms                                      | read,write |
| social_profile_types                      | social_profile_types                     | read       |
| source_types                              | source_types                             | read,write |
| users                                     | users                                    | read       |
| veteran_statuses                          | veteran_statuses                         | read       |
| workflow_stages                           | workflow_stages                          | read       |
| workflows                                 | workflows                                | read       |

The following objects support pagination using the "per_page" parameter for limit the page and "scroll_id" to fetch the next page
- form_templates
- forms
- people
- person_events
- scorecards
- sms
- countries
- jobs
- scorecards/scorecard_templates
- email_tracking
- schedule_items

The following objects support pagination using the "per_page" parameter for limit the page and "page" to fetch the next page
- countries
- jobs

The following objects support incremental read with "created_at_start" and "created_at_end" parameter
- email_tracking
- person_events
- sms

The following objects doesn't supports pagination and incremental read
- activity_types
- address_types
- bonus_payment_types
- bonus_types
- companies
- company_global_statuses
- company_types
- compensation_types
- currencies
- deal_workflows
- deals
- disability_statuses
- diversity_types
- dynamic_fields
- education_types
- email_types
- equity_types
- ethnicities
- fee_types
- genders
- job_categories
- job_contact_types
- job_owner_types
- job_statuses
- job_types
- person_global_statuses
- person_lists
- person_share_field_types
- person_types
- phone_types
- placements
- pronouns
- question_types
- scorecards/scorecard_recommendation_types
- scorecards/scorecard_types
- scorecards/scorecard_visibility_types
- seniority_levels
- social_profile_types
- source_types
- users
- veteran_statuses
- workflow_stages
- workflows

Endpoints supports create operation
- deals
- sms

Endpoints supports create, update operation
- people
- companies

Endpoints supports create, update and delete operation
- forms
- jobs
- person_events
- placements
- scorecards
- scorecard_templates
- source_types
