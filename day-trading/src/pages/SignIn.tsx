import styled from '@emotion/styled'
import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'

export default function SignIn() {
    return (
        <div>
            <SignBackground>
                <Card>
                    <SignInHeader>Sign In to Day-trades</SignInHeader>
                    <FieldForm>
                        <InputLabel>Username</InputLabel>
                        <SignField></SignField>
                        <InputLabel>Password</InputLabel>
                        <SignField type='password'></SignField>
                    </FieldForm>

                    <BigBlackButton>Sign in</BigBlackButton>
                    <LinkContainer>
                        <SignInText>New to Day-trades?</SignInText>
                        <SignInLink to='/'>Create an account</SignInLink>
                    </LinkContainer>
                </Card>
            </SignBackground>
        </div>
    )
}
