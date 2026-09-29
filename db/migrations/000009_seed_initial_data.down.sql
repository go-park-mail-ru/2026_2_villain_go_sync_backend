DELETE
FROM app_user
WHERE email IN (
                'employer@example.com',
                'seeker@example.com'
    );
