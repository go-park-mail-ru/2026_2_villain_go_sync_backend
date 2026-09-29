DROP TRIGGER IF EXISTS app_user_role_change_check ON app_user;
DROP TRIGGER IF EXISTS seeker_profile_role_check ON seeker_profile;
DROP TRIGGER IF EXISTS employer_profile_role_check ON employer_profile;

DROP FUNCTION IF EXISTS check_app_user_role_change();
DROP FUNCTION IF EXISTS check_profile_role();

DROP TABLE IF EXISTS seeker_profile;
DROP TABLE IF EXISTS employer_profile;
