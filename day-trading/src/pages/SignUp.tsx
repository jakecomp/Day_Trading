import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { HeaderContainer, InputContainer, LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { useForm } from 'react-hook-form'
import useSWR from 'swr'
import { useState } from 'react'

export const SignUp = () => {
    interface signForm {
        username: string
        password: string
    }

    const { register, handleSubmit } = useForm<signForm>({ mode: 'onSubmit' })
    const RetrieveData = (data: signForm) => {
        console.log(data)

        const report = {
            username: data.username,
            password: data.password,
        }

        try {
            fetch('http://localhost:8000/signup', {
                method: 'POST',
                mode: 'no-cors',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            }).then((response) => response.json)
        } catch (error) {
            console.error(error)
        }
    }

    return (
        <div>
            <SignBackground>
                <Card>
                    <HeaderContainer>
                        <SignInHeader>Welcome to Day-trades</SignInHeader>
                        <LinkContainer>
                            <SignInText>Already have an account?</SignInText>
                            <SignInLink to='/signin'>Sign in</SignInLink>
                        </LinkContainer>
                    </HeaderContainer>

                    <FieldForm onSubmit={handleSubmit(RetrieveData)}>
                        <InputContainer>
                            <InputLabel>Username</InputLabel>
                            <SignField {...register('username')}></SignField>
                            <InputLabel>Password</InputLabel>
                            <SignField
                                type='password'
                                placeholder='Must be at least 6 characters'
                                {...register('password')}
                            ></SignField>
                        </InputContainer>
                        <BigBlackButton>Create account</BigBlackButton>
                    </FieldForm>
                </Card>
            </SignBackground>
        </div>
    )
}
