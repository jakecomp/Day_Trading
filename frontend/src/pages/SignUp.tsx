/* eslint-disable prettier/prettier */
import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { HeaderContainer, InputContainer, LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { useForm } from 'react-hook-form'
import { SimpleLink } from '../components/atoms/links'
import { SignInPopUp } from '../components/popups/signinpopup'
import { Header3 } from '../components/atoms/fonts'
import { Component, useEffect, useRef, useState } from 'react'



export const SignUp = () => {

    const USER_REGEX = /^[A-z][A-z0-9-_]{3,23}$/;
    const PWD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%]).{8,24}$/;

    const [buttonPopup, setButtonPopup] = useState(false)

    interface signForm {
        username: string
        password: string
        confirmPassword: string
    }

    const { register, handleSubmit, watch, formState: {errors}, trigger } = useForm<signForm>({ mode: 'onSubmit' })
    const RetrieveData = (data: signForm) => {
        console.log(data)
        console.log(watch('username'))

        const report = {
            username: data.username,
            password: data.password,
        }

        let socket: WebSocket

        try {
            fetch('http://10.9.0.4:8000/signup', {
                method: 'POST',
                mode: 'no-cors',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = () => {
                        socket.send('Hi Hi Server')
                        console.log('Websocket Client Connected')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })

            // .then((ws = new WebSocket('ws://localhost:8000/ws?token=' + response)))
        } catch (error) {
            console.error(error)
        }
    }

    return (
        <section>
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
                            <SignField required={true} autoComplete='off'{...register('username', {required: "Username is required!", pattern: {value: /^[A-z][A-z0-9-_]{3,23}$/, message: "Invalid username"}})}></SignField>
                            {errors.username && (<Header3>{errors.username.message}</Header3>)}
                            <InputLabel>Password</InputLabel>
                            <SignField
                                type='password'
                                placeholder='Must be at least 6 characters'
                                {...register('password', {required: true})}
                            ></SignField>
                            <InputLabel>Confirm Password</InputLabel>
                            <SignField
                                type='password'
                                placeholder=''
                                required={true}
                                {...register('confirmPassword', {validate: value => value === watch("password", "") || "The passwords do not match"})}
                                autoComplete='off'
                            ></SignField>
                        </InputContainer>
                        <BigBlackButton onClick={() => setButtonPopup(true)}>Create Account</BigBlackButton>
                    </FieldForm>
                </Card>
            </SignBackground>
            <SignInPopUp trigger = {buttonPopup}>
                <Header3>Account Created!</Header3>
                <SimpleLink to='/home'>
                    <BigBlackButton style={{width: '300px'}}>Go to Home</BigBlackButton>
                </SimpleLink>
            </SignInPopUp>
        </section>
            
        
    )
}
